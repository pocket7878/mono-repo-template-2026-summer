// Package repositoryentityreconstruct prevents repositories from bypassing entity constructors.
package repositoryentityreconstruct

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const reconstructPrefix = "Reconstruct"

var Analyzer = &analysis.Analyzer{
	Name: "repository_entity_reconstruct",
	Doc:  "require entity.Reconstruct* when repositories recreate entities",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !isTargetRepositoryPackage(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.CallExpr:
				checkCallExpr(pass, n)
			case *ast.CompositeLit:
				checkCompositeLit(pass, n)
			}
			return true
		})
	}
	return nil, nil
}

func checkCallExpr(pass *analysis.Pass, call *ast.CallExpr) {
	if isBuiltinNewOnEntity(pass, call) {
		pass.Reportf(call.Pos(), "repository layer must not instantiate entity types directly: new(entity.*) is forbidden; use entity.Reconstruct*")
		return
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	fn, ok := pass.TypesInfo.Uses[selector.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil || !isEntityPackage(fn.Pkg().Path()) || strings.HasPrefix(fn.Name(), reconstructPrefix) || !functionReturnsEntityType(fn.Type()) {
		return
	}
	pass.Reportf(call.Pos(), "repository layer must use entity.Reconstruct* for entity construction: entity.%s is forbidden", fn.Name())
}

func checkCompositeLit(pass *analysis.Pass, lit *ast.CompositeLit) {
	if isEntityType(pass.TypesInfo.TypeOf(lit.Type)) {
		pass.Reportf(lit.Pos(), "repository layer must not instantiate entity types directly: entity composite literal is forbidden; use entity.Reconstruct*")
	}
}

func isBuiltinNewOnEntity(pass *analysis.Pass, call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "new" && len(call.Args) == 1 && isEntityType(pass.TypesInfo.TypeOf(call.Args[0]))
}

func functionReturnsEntityType(fnType types.Type) bool {
	signature, ok := fnType.(*types.Signature)
	if !ok || signature.Results() == nil {
		return false
	}
	for i := 0; i < signature.Results().Len(); i++ {
		if isEntityType(signature.Results().At(i).Type()) {
			return true
		}
	}
	return false
}

func isEntityType(t types.Type) bool {
	if t == nil {
		return false
	}
	switch tt := t.(type) {
	case *types.Named:
		return tt.Obj() != nil && tt.Obj().Pkg() != nil && isEntityPackage(tt.Obj().Pkg().Path())
	case *types.Alias:
		return isEntityType(tt.Rhs())
	case *types.Pointer:
		return isEntityType(tt.Elem())
	case *types.Array:
		return isEntityType(tt.Elem())
	case *types.Slice:
		return isEntityType(tt.Elem())
	case *types.Map:
		return isEntityType(tt.Key()) || isEntityType(tt.Elem())
	case *types.Chan:
		return isEntityType(tt.Elem())
	}
	return false
}

func isTargetRepositoryPackage(pkgPath string) bool {
	return strings.HasSuffix(pkgPath, "/internal/repository") || strings.Contains(pkgPath, "/internal/repository/")
}

func isEntityPackage(pkgPath string) bool { return strings.HasSuffix(pkgPath, "/internal/entity") }

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	positionFile := pass.Fset.File(file.Pos())
	return positionFile != nil && strings.HasSuffix(positionFile.Name(), "_test.go")
}

# 2026-10-21 の公式セキュリティ版。nixpkgs 全体の更新は避ける。
final: prev:
let
  base = if builtins.hasAttr "paths" prev.nodejs_22
    then prev.nodejs-slim_22 else prev.nodejs_22;
  patched = base.overrideAttrs (old: {
    version = "22.23.4";
    # 修正済みの同梱 Brotli/c-ares/OpenSSL/SQLite を使う。
    configureFlags = builtins.filter (flag:
      !(builtins.any (name: prev.lib.hasPrefix ("--shared-" + name) flag)
        [ "brotli" "cares" "openssl" "sqlite" ])
      && flag != "--openssl-use-def-ca-store"
    ) old.configureFlags;
    src = prev.fetchurl {
      url = "https://nodejs.org/dist/v22.23.4/node-v22.23.4.tar.xz";
      sha256 = "7b9adc82d612d0a2a863d96c46de2292552e250cee51956ebbe0146c42aacae9";
    };
  });
in {
  nodejs-slim_22 = patched;
  nodejs-slim = patched;
  nodejs_22 = if builtins.hasAttr "paths" prev.nodejs_22
    then prev.nodejs_22.override { nodejs-slim = patched; }
    else patched;
  nodejs = final.nodejs_22;
}

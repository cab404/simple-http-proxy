{
    inputs = {
        nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
        utils.url = "github:numtide/flake-utils";
    };

    outputs = {self, nixpkgs, utils}:
    let out = system:
    let pkgs = nixpkgs.legacyPackages."${system}";
    in {

        devShell = pkgs.mkShell {
            buildInputs = with pkgs; [
                go
            ];
        };

        defaultPackage = pkgs.buildGoPackage {
            pname = "simple-http-proxy";
            version = "0.1.0";
            goPackagePath = "github.com/cab404/simple-http-proxy";
            src = ./.;
        };

        defaultApp = utils.lib.mkApp {
            drv = self.defaultPackage."${system}";
        };

    }; in with utils.lib; eachSystem defaultSystems out;

}

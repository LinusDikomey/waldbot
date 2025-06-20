{
  description = "Waldbot";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    nixpkgsFor = system: import nixpkgs {inherit system;};
  in {
    packages = forAllSystems (system: let
      pkgs = nixpkgsFor system;
    in {
      waldbot = pkgs.buildGoModule {
        pname = "waldbot";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-/SSobuox1+YG2tWLMfbO1cFp3E9kd/5Mw11GOWA4SR0=";
        buildInputs = [];
      };
      default = self.packages.${system}.waldbot;
    });
    devShells = forAllSystems (system: let
      pkgs = nixpkgsFor system;
    in {
      default = pkgs.mkShell {
        inputsFrom = [self.packages.${system}.default];
        nativeBuildInputs = with pkgs; [gopls];
      };
    });
  };
}

self:
{ config
, lib
, pkgs
, ...
}:
let
  inherit (pkgs.stdenv.hostPlatform) system;
  package = self.packages.${system}.default;
  cfg = config.services.protrans;
  inherit (lib) mkEnableOption mkIf;
in
{
  options.services.protrans = {
    enable = mkEnableOption "enable protrans service";
  };

  config = mkIf cfg.enable {
    systemd.user.services.protrans = {
      Unit = {
        Description = "ProTrans service";
        After = "network.target";
      };

      Service = {
        ExecStart = "${package}/bin/protrans";
        Restart = "on-abnormal";
      };

      Install = {
        WantedBy = [ "graphical-session.target" ];
      };
    };
  };
}

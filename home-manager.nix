self:
{ config
, lib
, pkgs
, ...
}:
let
  inherit (pkgs) writeTextFile;
  inherit (pkgs.stdenv.hostPlatform) system;
  package = self.packages.${system}.default;
  cfg = config.services.protrans;
  inherit (lib) mkEnableOption mkOption mkIf types;
  mkStringOption = description: default: mkOption {
    inherit description default;
    type = types.str;
  };
  mkIntOption = description: default: mkOption {
    inherit description default;
    type = types.int;
  };
in
{
  options.services.protrans = {
    enable = mkEnableOption "enable protrans service";
    configuration = {
      logLevel = mkOption {
        type = types.enum [ "TRACE" "DEBUG" "INFO" "WARN" "ERROR" "FATAL" "PANIC" ];
        description = "Set the log level";
        default = "INFO";
      };
      transmission = {
        host = mkStringOption "Host for Transmission" "localhost";
        port = mkIntOption "Port to use to communicate with Transmission" 9091;
        username = mkStringOption "User for transmission" "";
        password = mkStringOption "Password for transmission" "";
      };
      nat = {
        gateway = mkStringOption "IP Address of the Gateway for NAT" "10.2.0.1";
        portLifeTime = mkIntOption "Port Lifetime in seconds" 600;
      };
    };
  };

  config = mkIf cfg.enable {
    systemd.user.services.protrans =
      let
        configFile = writeTextFile {
          name = "protrans-config.yaml";
          text = lib.generators.toYAML { } {
            inherit (cfg.configuration) transmission;
            log_level = cfg.configuration.logLevel;
            nat = {
              inherit (cfg.configuration.nat) gateway;
              port_lifetime = cfg.configuration.nat.portLifeTime;
            };
          };
        };
      in
      {
        Unit = {
          Description = "ProTrans service";
          After = "network.target";
        };

        Service = {
          ExecStart = "${package}/bin/protrans ${configFile}";
          Restart = "on-abnormal";
        };

        Install = {
          WantedBy = [ "graphical-session.target" ];
        };
      };
  };
}

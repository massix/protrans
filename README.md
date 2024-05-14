# Protrans: Open NAT Ports and Configure Transmission Automatically

## Introduction

Protrans is a user-friendly, open-source Go application designed to streamline
the process of opening NAT ports using the natpmp library and automatically
configuring them for the Transmission Torrent client. It offers the flexibility
to be used either as a standalone tool or seamlessly integrated as a service
within systemd.

## Features

- Effortless NAT Port Opening: Protrans leverages the natpmp library to
establish communication with your router, allowing you to effortlessly open the
necessary ports for Transmission to operate effectively.

- Automatic Transmission Configuration: Protrans intelligently configures
Transmission's port settings, ensuring seamless integration and optimal
performance without manual intervention.

- Standalone or Systemd Service: Protrans caters to both standalone and systemd
service deployment models. Choose the approach that best suits your workflow
and system management preferences.

- Compatibility: Protrans is specifically designed to work with ProtonVPN.
However, it includes a configuration file that may potentially work with other
VPN providers, although this functionality is untested.


## Using with Home Manager

If you use either NixOS or nixpkgs with home manager, you can import the whole
project as a flake into your configuration and use it as a HM Module like you
would do for any other NixOS module.

## Installation

**Prerequisites:**

Go version 1.17 or later (https://go.dev/doc/install)

**Standalone Usage:**

Clone the Protrans repository:

```bash
git clone https://github.com/massix/protrans.git
```

Navigate to the project directory:

```bash
cd protrans
```

Build the Protrans executable:

```bash
go build -o protrans cmd/protrans/main.go
```

**Run Protrans:**

```bash
./protrans
```

## Systemd Service:

Create a systemd service file (`protrans.service`) in a directory accessible by
your systemd configuration (e.g., /etc/systemd/system/). Here's a sample
service file:

```
[Unit]
Description=Protrans - NAT Port Opener and Transmission Configurator
After=network.target

[Service]
User=your-username  # Replace with your system user
Group=your-group    # Replace with your system group
WorkingDirectory=/path/to/protrans  # Replace with the path to your Protrans directory
ExecStart=/path/to/protrans  # Replace with the path to the Protrans executable

[Install]
WantedBy=multi-user.target
```

Reload the systemd configuration:

```bash
sudo systemctl daemon-reload
```

**Enable and start the Protrans service:**

```bash
sudo systemctl enable protrans.service
sudo systemctl start protrans.service
```

**Usage (Standalone)**

Once you've built the Protrans executable, simply run it from the command line:

```bash
./protrans
```

Protrans will automatically open the necessary NAT ports and configure
Transmission for optimal operation.

**Usage (Systemd Service)**

After configuring and enabling the systemd service, Protrans will automatically
start at system boot and continue running in the background, ensuring your
Transmission client remains properly configured.

## Contributing

We welcome contributions from the community! Feel free to fork the repository,
make changes, and submit pull requests.

## License

Protrans is licensed under the MIT License.

## Contact

For any questions or feedback, feel free to create an issue on the project's
GitHub repository.



# `portmaster-plugin-dnscrypt`

**DISCLAIMER**: This is not an official Safing product!  

This repository provides a [Safing Portmaster](https://github.com/safing/portmaster) resolver plugin that resolves DNS queries using a DNSCrypt server.

**Warning**: This repository is based on the experimental Portmaster Plugin System which is available in [safing/portmaster#834](https://github.com/safing/portmaster/pull/834) but has not been merged and released yet.

## Installation

### Using the install command

This plugin uses the `cmds.InstallCommand()` from the portmaster plugin framework so installation is as simple as:

```bash
go build .
sudo ./portmaster-plugin-dnscrypt install --data /opt/safing/portmaster
```

### Manual Installation

To manually install the plugin follow these steps:

1. Build the plugin from source code: `go build .`
2. Move the plugin `/opt/safing/portmaster/plugins/portmaster-plugin-dnscrypt`
3. Edit `/opt/safing/portmaster/plugins.json` to contain the following content:

   ```
   [
        {
            "name": "portmaster-plugin-dnscrypt",
            "types": [
                "decider"
            ],
        }
   ]
   ```

## Configuration

**Important**: Before being able to use plugins in the Portmaster you must enable the "Plugin System" in the global settings page. Note that this setting is still marked as "Experimental" and "Developer-Only" so you'r Portmaster needs the following settings adjusted to even show the "Plugin System" setting:

 - [Developer Mode](https://docs.safing.io/portmaster/settings#core/devMode)
 - [Feature Stability](https://docs.safing.io/portmaster/settings#core/releaseLevel)

This plugin registers a new setting `"plugins/portmaster-plugin-dnscrypt/dnscryptServer"` in the Portmaster so you can just paste the server-stamp of the DNSCrypt server you want to use there.

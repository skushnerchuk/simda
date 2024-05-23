// build +darwin

package config

import "github.com/spf13/viper"

func setDefaults() {
	viper.SetDefault("metrics.load_avg", true)
	viper.SetDefault("metrics.cpu_avg", false)
	viper.SetDefault("metrics.disk_io", false)
	viper.SetDefault("metrics.disk_usage", false)
	viper.SetDefault("metrics.net_connections", false)
	viper.SetDefault("metrics.net_connections_states", false)
	viper.SetDefault("metrics.net_top_by_protocol", false)
	viper.SetDefault("metrics.net_top_by_connection", false)

	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", "50051")
	viper.SetDefault("log_level", "DEBUG")
	viper.SetDefault("system.proc", "/proc")
	viper.SetDefault("system.sys", "/sys")
	viper.SetDefault("system.dev", "/dev")
	viper.SetDefault("system.run", "/run")
	viper.SetDefault("system.tcp", "/proc/net/tcp")
	viper.SetDefault("system.tcp6", "/proc/net/tcp6")
	viper.SetDefault("system.udp", "/proc/net/udp")
	viper.SetDefault("system.udp6", "/proc/net/udp6")
	viper.SetDefault("system.procMountInfo", "")
	viper.SetDefault("system.interface", "any")
}

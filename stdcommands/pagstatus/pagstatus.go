package pagstatus

import (
	"fmt"
	"runtime"
	"time"

	"github.com/jonas747/dcmd/v3"
	"github.com/jonas747/discordgo"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

var Command = &commands.YAGCommand{
	Cooldown:    5,
	CmdCategory: commands.CategoryDebug,
	Name:        "Pagstatus",
	Aliases:     []string{"status"},
	Description: "Shows PAGSTDB status, version, uptime, memory stats, etc...",
	RunInDM:     true,
	RunFunc:     cmdFuncYagStatus,
}

var logger = common.GetFixedPrefixLogger("yagstatuc_cmd")

func cmdFuncYagStatus(data *dcmd.Data) (interface{}, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	servers, _ := common.GetJoinedServerCount()

	sysMem, err := mem.VirtualMemory()
	sysMemStats := ""
	if err == nil {
		sysMemStats = fmt.Sprintf("%dMB (%.0f%%), %dMB", sysMem.Used/1000000, sysMem.UsedPercent, sysMem.Total/1000000)
	} else {
		sysMemStats = "Failed collecting mem stats"
		logger.WithError(err).Error("Failed collecting memory stats")
	}

	sysLoad, err := load.Avg()
	sysLoadStats := ""
	if err == nil {
		sysLoadStats = fmt.Sprintf("%.2f, %.2f, %.2f", sysLoad.Load1, sysLoad.Load5, sysLoad.Load15)
	} else {
		sysLoadStats = "Failed collecting"
		logger.WithError(err).Error("Failed collecting load stats")
	}

	uptime := time.Since(bot.Started)
	hostUptimeSeconds, _ := host.Uptime()
	hostUptimeDays := hostUptimeSeconds / (60 * 60 * 24)
	hostUptimeHours := (hostUptimeSeconds - (hostUptimeDays * 60 * 60 * 24)) / (60 * 60)
	hostUptimeMinutes := ((hostUptimeSeconds - (hostUptimeDays * 60 * 60 * 24)) - (hostUptimeHours * 60 * 60)) / 60
	hostUptime := fmt.Sprintf("%d days, %d hours, %d minutes", hostUptimeDays, hostUptimeHours, hostUptimeMinutes)
	allocated := float64(memStats.Alloc) / 1000000

	numGoroutines := runtime.NumGoroutine()

	botUser := common.BotUser

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    botUser.Username,
			IconURL: discordgo.EndpointUserAvatar(botUser.ID, botUser.Avatar),
		},
		Title: "PAGSTDB Status, version " + common.VERSION,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{Name: "Servers", Value: fmt.Sprint(servers), Inline: true},
			&discordgo.MessageEmbedField{Name: "Go Version", Value: runtime.Version(), Inline: true},
			&discordgo.MessageEmbedField{Name: "Uptime", Value: fmt.Sprintf("Bot: %s\nHost: %s", common.HumanizeDuration(common.DurationPrecisionSeconds, uptime), hostUptime), Inline: true},
			&discordgo.MessageEmbedField{Name: "Goroutines", Value: fmt.Sprint(numGoroutines), Inline: true},
			&discordgo.MessageEmbedField{Name: "GC Pause Fraction", Value: fmt.Sprintf("%.3f%%", memStats.GCCPUFraction*100), Inline: true},
			&discordgo.MessageEmbedField{Name: "Process Mem (alloc, sys, freed)", Value: fmt.Sprintf("%.1fMB, %.1fMB, %.1fMB", float64(memStats.Alloc)/1000000, float64(memStats.Sys)/1000000, (float64(memStats.TotalAlloc)/1000000)-allocated), Inline: true},
			&discordgo.MessageEmbedField{Name: "System Mem (used, total)", Value: sysMemStats, Inline: true},
			&discordgo.MessageEmbedField{Name: "System Load (1, 5, 15)", Value: sysLoadStats, Inline: true},
		},
	}

	for _, v := range common.Plugins {
		if cast, ok := v.(PluginStatus); ok {
			started := time.Now()
			name, val := cast.Status()
			if name == "" || val == "" {
				continue
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: v.PluginInfo().Name + ": " + name, Value: val, Inline: true})
			elapsed := time.Since(started)
			logger.Println("Took ", elapsed.Seconds(), " to gather stats from ", v.PluginInfo().Name)
		}
	}

	return embed, nil
	// return &commandsystem.FallbackEmebd{embed}, nil
}

type PluginStatus interface {
	Status() (string, string)
}

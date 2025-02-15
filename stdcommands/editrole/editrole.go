package editrole

import (
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/moderation"
	"github.com/jonas747/dcmd/v4"
	"github.com/jonas747/discordgo/v2"
	"golang.org/x/image/colornames"
)

var Command = &commands.YAGCommand{
	CmdCategory:     commands.CategoryTool,
	Name:            "EditRole",
	Aliases:         []string{"ERole"},
	Description:     "Edits a role",
	LongDescription: "Requires the manage roles permission and the bot and your highest role being above the edited role. Role permissions follow discord standard encoding can can be calculated [here](https://discordapp.com/developers/docs/topics/permissions)",
	RequiredArgs:    1,
	Arguments: []*dcmd.ArgDef{
		{Name: "Role", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "name", Help: "Role name - String", Type: dcmd.String, Default: ""},
		{Name: "color", Help: "Role color - Either hex code or name", Type: dcmd.String, Default: ""},
		{Name: "mention", Help: "Role Mentionable - 1 for true 0 for false", Type: &dcmd.IntArg{Min: 0, Max: 1}},
		{Name: "hoist", Help: "Role Hoisted - 1 for true 0 for false", Type: &dcmd.IntArg{Min: 0, Max: 1}},
		{Name: "perms", Help: "Role Permissions - 0 to 2147483647", Type: &dcmd.IntArg{Min: 0, Max: 2147483647}},
	},
	RunFunc:            cmdFuncEditRole,
	GuildScopeCooldown: 30,
}

func cmdFuncEditRole(data *dcmd.Data) (interface{}, error) {
	cID := data.ChannelID
	if ok, err := bot.AdminOrPermMS(data.GuildData.GS.ID, cID, data.GuildData.MS, discordgo.PermissionManageRoles); err != nil {
		return "Failed checking perms", err
	} else if !ok {
		return "You need manage roles perms to use this command", nil
	}

	roleS := data.Args[0].Str()
	role := moderation.FindRole(data.GuildData.GS, roleS)

	if role == nil {
		return "No role with the Name or ID`" + roleS + "` found", nil
	}

	//data.GuildData.GS.RLock()
	if !bot.IsMemberAboveRole(data.GuildData.GS, data.GuildData.MS, role) {
		//data.GuildData.GS.RUnlock()
		return "Can't edit roles above you", nil
	}
	//data.GuildData.GS.RUnlock()

	change := false

	name := role.Name
	if n := data.Switch("name").Str(); n != "" {
		name = limitString(n, 100)
		change = true
	}
	color := role.Color
	if c := data.Switch("color").Str(); c != "" {
		if data.Source == dcmd.TriggerSourceDM {
			return nil, errors.New("Cannot use role color edit in custom commands to prevent api abuse")
		}
		parsedColor, ok := ParseColor(c)
		if !ok {
			return "Unknown color: " + c + ", can be either hex color code or name for a known color", nil
		}
		color = parsedColor
		change = true
	}
	mentionable := role.Mentionable
	if m := data.Switch("mention"); m != nil {
		mentionable = m.Bool()
		change = true
	}
	hoisted := role.Hoist
	if h := data.Switch("hoist"); h != nil {
		hoisted = h.Bool()
		change = true
	}
	perms := role.Permissions
	if p := data.Switch("perms"); p != nil {
		perms = p.Int64()
		change = true
	}

	if change {
		_, err := common.BotSession.GuildRoleEdit(data.GuildData.GS.ID, role.ID, name, color, hoisted, perms, mentionable)
		if err != nil {
			return nil, err
		}
	}

	_, err := common.BotSession.ChannelMessageSendComplex(cID, &discordgo.MessageSend{
		Content:         fmt.Sprintf("__**Edited Role(%d) properties to :**__\n\n**Name **: `%s`\n**Color **: `%d`\n**Mentionable **: `%t`\n**Hoisted **: `%t`\n**Permissions **: `%d`", role.ID, name, color, mentionable, hoisted, perms),
		AllowedMentions: discordgo.AllowedMentions{},
	})

	if err != nil {
		return nil, err
	}

	return nil, err
}

func ParseColor(raw string) (int, bool) {
	if strings.HasPrefix(raw, "#") {
		raw = raw[1:]
	}

	// try to parse as hex color code first
	parsed, err := strconv.ParseInt(raw, 16, 32)
	if err == nil {
		temp := int(parsed)
		if temp > 16777215 {
			temp = 16777215
		}
		return temp, true
	}

	// look up the color code table
	for _, v := range colornames.Names {
		if strings.EqualFold(v, raw) {
			cStruct := colornames.Map[v]

			color := (int(cStruct.R) << 16) | (int(cStruct.G) << 8) | int(cStruct.B)
			return color, true
		}
	}

	return 0, false
}

// limitstring cuts off a string at max l length, supports multi byte characters
func limitString(s string, l int) string {
	if len(s) <= l {
		return s
	}

	lastValidLoc := 0
	for i, _ := range s {
		if i > l {
			break
		}
		lastValidLoc = i
	}

	return s[:lastValidLoc]
}

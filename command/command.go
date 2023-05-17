package command

import (
	"fmt"
	"log"
	"waldbot/data"
	"waldbot/date"

	"github.com/bwmarrin/discordgo"
)

type Query struct {
    member *discordgo.Member
    dateCondition date.DateCondition
    selfUser bool
}

var (
    ZEITRAUM_OPTION = discordgo.ApplicationCommandOption {
        Type: discordgo.ApplicationCommandOptionString,
        Name: "Zeitraum",
        Description: "z.B. daily', 'weekly', '1.2.2021', '15.4.2020-17.6.2020'",
        Required: false,
    }

    NUTZER_OPTION = discordgo.ApplicationCommandOption {
        Type: discordgo.ApplicationCommandOptionUser,
        Name: "Nutzer",
        Description: "Nutzer, dessen Zeit angezeigt werden soll",
        Required: false,
    }
)

type Options struct {
    zeitraum, nutzer bool
}

type SlashCommand struct {
    name string
    description string
    response func(Query) string
    options Options
}

var (
	SlashCommands = []SlashCommand {
        {
            name: "time",
			description: "Zeigt Sprachchatzeit, Rang und Lieblingskanäle an",
            response: timeResponse,
            options: Options { zeitraum: true, nutzer: true },
        },
        {
            name: "hours",
            description: "Generiert ein Diagram mit der Sprachchat-Zeitverteilung über einen Zeitraum",
            response: hoursResponse,
            options: Options { zeitraum: true, nutzer: true },
        },
        {
            name: "channels",
            description: "Zeigt ein Tortendiagramm mit der Verteilung der genutzten Sprachkanäle",
            //response: 
            options: Options { zeitraum: true, nutzer: true },
        },
	}
    registeredCommands = make([]*discordgo.ApplicationCommand, len(SlashCommands))
)

func RegisterCommands(dc *discordgo.Session) {
	for i, cmd := range SlashCommands {
        options := make([]*discordgo.ApplicationCommandOption, 0)
        if cmd.options.nutzer {
            options = append(options, &NUTZER_OPTION)
        }
        if cmd.options.zeitraum {
            options = append(options, &ZEITRAUM_OPTION)
        }
		fmt.Println("Adding command:", cmd.name, ", stateId:", dc.State.User.ID)
		command, err := dc.ApplicationCommandCreate(dc.State.User.ID, "", &discordgo.ApplicationCommand {
            Name: cmd.name,
            Description: cmd.description,
            Options: options,
        })
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", cmd.name, err)
		}
        registeredCommands[i] = command
    }
}

func UnregisterCommands(dc *discordgo.Session) {
    for _, v := range registeredCommands {
        err := dc.ApplicationCommandDelete(dc.State.User.ID, "", v.ID)
        if err != nil {
            log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
        }
    }
}

func memberValue(guildID string, o *discordgo.ApplicationCommandInteractionDataOption) *discordgo.Member {
	userID := o.StringValue()
	if userID == "" {
		return nil
	}

	m, err := data.Dc.State.Member(guildID, userID)
	if err != nil {
		return nil
	}

	return m
}

func parseOptions(
    options []*discordgo.ApplicationCommandInteractionDataOption,
    member *discordgo.Member,
    validOptions Options,
) (Query, string) {
    query := Query {
        member: member,
        dateCondition: date.AllTimeCondition,
        selfUser: true,
    }
    for _, option := range options {
        switch option.Name {
        case "nutzer":
            query.member = memberValue(member.GuildID, option)
            if query.member == nil {
                return query, "invalider Nutzer"
            }
        case "zeitraum":
            parsedDateCondition, status := data.ParseDateCondition(option.StringValue(), date.AllTimeCondition)
            if status != data.ParseSuccess {
                return query, "invalider Zeitraum"
            }
            query.dateCondition = parsedDateCondition
        default:
            return query, "invalides Argument " + option.Name
        }

    }

    return query, ""
}

func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
    for _, cmd := range SlashCommands {
        if cmd.name == i.ApplicationCommandData().Name {
            query, err := parseOptions(i.ApplicationCommandData().Options, i.Member, cmd.options)
            var content string
            if err != "" {
                content = "Fehler: " + err
            } else {
                content = cmd.response(query)
            }
            
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse {
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData {
                    Content: content,
                },
            })
        }
    }
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData {
            Content: "Fehler: der Befehl '" + i.ApplicationCommandData().Name + "' existiert nicht :(",
        },
    })
}


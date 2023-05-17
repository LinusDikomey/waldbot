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
        Name: "zeitraum",
        Description: "z.B. daily', 'weekly', 'yearly', '1.2.2021', '15.4.2020-17.6.2020'",
        // choices geht nicht, da dann andere Werte nicht mehr akzeptiert werden
        /*
        Choices: []*discordgo.ApplicationCommandOptionChoice {
            {
                Name: "daily",
                Value: "daily",
            },
            {
                Name: "weekly",
                Value: "weekly",
            },
            {
                Name: "yearly",
                Value: "yearly",
            },
            {
                Name: "1.1.2023-9.1.2023",
                Value: "1.1.2023-9.1.2023",
            },
        },
        */
        Required: false,
    }

    NUTZER_OPTION = discordgo.ApplicationCommandOption {
        Type: discordgo.ApplicationCommandOptionUser,
        Name: "nutzer",
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
    response func(Query) (string, *discordgo.File)
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
            response: hoursResponse, // TODO: channels command
            options: Options { zeitraum: true, nutzer: true },
        },
	}
    registeredCommands = make([]*discordgo.ApplicationCommand, len(SlashCommands))
)

const FALSE bool = false

func RegisterCommands(dc *discordgo.Session, guildID string) {
    // clean up previous commands that might still be hanging around
    oldCommands, err := dc.ApplicationCommands(dc.State.User.ID, "")
    if err == nil {
        for _, command := range oldCommands {
            dc.ApplicationCommandDelete(dc.State.User.ID, "", command.ID)
        }
    } else {
        fmt.Println("Failed to retrieve old application commands:", err)
    }
    oldGuildCommands, err := dc.ApplicationCommands(dc.State.User.ID, guildID)
    if err == nil {
        for _, command := range oldGuildCommands {
            dc.ApplicationCommandDelete(dc.State.User.ID, guildID, command.ID)
        }
    } else {
        fmt.Println("Failed to retrieve old application commands:", err)
    }


	for i, cmd := range SlashCommands {
        options := make([]*discordgo.ApplicationCommandOption, 0)
        if cmd.options.nutzer {
            options = append(options, &NUTZER_OPTION)
        }
        if cmd.options.zeitraum {
            options = append(options, &ZEITRAUM_OPTION)
        }
        fmt.Println("Adding command:", cmd.name, "UserID: ", dc.State.User.ID)
        dmPermission := false
		command, err := dc.ApplicationCommandCreate(dc.State.User.ID, "", &discordgo.ApplicationCommand {
            Name: cmd.name,
            Description: cmd.description,
            Options: options,
            DMPermission: &dmPermission,
        })
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.name, err)
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

func memberValue(s *discordgo.Session, guildID string, o *discordgo.ApplicationCommandInteractionDataOption) *discordgo.Member {
	user := o.UserValue(nil)
    if user == nil {
        log.Println("User was nil")
        return nil
    }
	m, err := s.State.Member(guildID, user.ID)
	if err != nil {
        log.Println("User to member err: ", err)
		return nil
	}

	return m
}

func parseOptions(
    options []*discordgo.ApplicationCommandInteractionDataOption,
    member *discordgo.Member,
    guildID string,
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
            query.member = memberValue(data.Dc, guildID, option)
            if query.member == nil {
                return query, "invalider Nutzer"
            }
            if query.member.User.ID != member.User.ID {
                query.selfUser = false
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
            query, err := parseOptions(i.ApplicationCommandData().Options, i.Member, i.GuildID, cmd.options)
            var content string
            files := make([]*discordgo.File, 0)
            if err != "" {
                content = "Fehler: " + err
            } else {
                var file *discordgo.File
                content, file = cmd.response(query)
                if file != nil {
                    files = append(files, file)
                }
            }
            
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse {
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData {
                    Content: content,
                    Files: files,
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


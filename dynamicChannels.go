package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func onChannelConnect(s *discordgo.Session, m *discordgo.Connect) {
	
}

func onChannelDisconnect(s *discordgo.Session, m *discordgo.Disconnect) {

}

func updateDynamicChannels() {
	
	removeElem := func(a *[]string, i int) {
		copy((*a)[i:], (*a)[i+1:]) // Shift a[i+1:] left one index.
		(*a)[len(*a)-1] = ""     // Erase last element (write zero value).
		*a = (*a)[:len(*a)-1]
	}

	offset := config.DynamicChannels.Position
	counts := countUsersInChannels()
	
	for _, channelName := range config.DynamicChannels.Channels {
		if instanceChannels, ok := data.DynamicChannels[channelName]; ok {
			instanceIndex := 0
			removedInstances := []string {}

			for i, channelId := range instanceChannels {
				lastInstance := i == len(instanceChannels)-1 
				if _, ok := counts[channelId]; ok {
					channel, _ := dc.State.Channel(channelId)
					instanceName := fmt.Sprintf("%v #%v", channelName, i+1)
					if channel.Name != instanceName {
						dc.ChannelEdit(channelId, instanceName)
					}
					instanceIndex++
				} else if !lastInstance { // 0 users in channel, delete unless last instance
					dc.ChannelDelete(channelId)
					removedInstances = append(removedInstances, channelId)
				}
			}
			for _, remove := range removedInstances {
				index := -1
				for i, search := range instanceChannels {
					if search == remove {
						index = i
					}
				}
				if index == -1 { log.Fatal("Was zur Hölle?") }
				removeElem(&instanceChannels, index)
			}
			// check wether new channel has to be created
			/*if _, ok := counts[instanceChannels[len(instanceChannels)-1]]; ok {
				instanceChannels = append(instanceChannels, createDynamicVoiceChannel(
					fmt.Sprintf("%v #%v",  , instanceIndex))
			}*/
			offset += len(instanceChannels)
		} else {
			createDynamicVoiceChannel(fmt.Sprintf("%v #1", channelName), offset)
			offset++
			data.DynamicChannels[channelName] = []string {}
		}
	}
}

func createDynamicVoiceChannel(name string, position int) string {
	channel, err := dc.GuildChannelCreateComplex(config.GuildId, discordgo.GuildChannelCreateData {
		Name: name,
		Type: discordgo.ChannelTypeGuildVoice,
		Bitrate: config.DynamicChannels.Bitrate,
		Position: position,
	})
	if err != nil {
		log.Fatal("Could not create dynamic voice channel:", channel)
	}
	return channel.ID
}

func countUsersInChannels() map[string]uint32 {
	members, err := dc.GuildMembers(config.GuildId, "", 1000)
	if err != nil {
		panic("Could not find Guild by id specified in config: " + fmt.Sprint(err))
	}
	counts := map[string]uint32 {}

	for _, member := range members {
		if member.User.Bot { continue }
		state, _ := dc.State.VoiceState(config.GuildId, member.User.ID)
		if _, ok := counts[state.ChannelID]; ok {
			counts[state.ChannelID] += 1
		} else {
			counts[state.ChannelID] = 1
		}
	}
	return counts
}
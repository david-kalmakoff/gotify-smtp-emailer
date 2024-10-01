package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "https://github.com/david-kalmakoff/gotify-smtp-emailer",
		Version:     "0.1.0",
		Author:      "David Kalmakoff",
		Description: "A plugin for sending smtp emails for incoming gotify/server messages.",
		License:     "MIT",
		Name:        "Gotify SMTP Emailer",
	}
}

// Plugin is the gotify plugin instance.
type Plugin struct {
	userCtx    plugin.UserContext
	msgHandler plugin.MessageHandler
	config     *Config
	enabled    bool
	connection *websocket.Conn
	done       chan bool
	err        error
}

type GotifyMessage struct {
	Id       uint32
	Appid    uint32
	Message  string
	Title    string
	Priority uint32
	Date     string
}

// Enable enables the plugin.
func (c *Plugin) Enable() error {
	err := c.config.IsValid()
	if err != nil {
		return fmt.Errorf("config is invalid: %w", err)
	}

	// start websocket connection
	c.connection, err = c.config.getWSConnection()
	if err != nil {
		return fmt.Errorf("could not get ws connection: %w", err)
	}

	c.done = make(chan bool)

	go func() {
		defer close(c.done)

		for {
			select {
			case <-c.done:
				return
			default:
				msg := GotifyMessage{}
				err := c.connection.ReadJSON(&msg)
				if err != nil {
					if _, ok := err.(*websocket.CloseError); ok {
						return
					}
					log.Printf("connection read error: %v", err)
					continue
				}

				// send message to smtp
				err = c.config.Smtp.Send(msg.Title, msg.Message, c.config.Environment == "development")
				if err != nil {
					log.Printf("smtp send error: %v", err)
				}
			}
		}
	}()

	if c.config.Environment == "development" {
		go func() {
			for {
				if !c.enabled {
					return
				}
				c.msgHandler.SendMessage(plugin.Message{
					Title:   "Test Message",
					Message: fmt.Sprintf("config: %#v ,error: %v", c.config, c.err),
				})
				time.Sleep(time.Second * 10)
			}
		}()
	}

	c.enabled = true

	return nil
}

// Disable disables the plugin.
func (c *Plugin) Disable() error {
	err := c.connection.Close()
	if err != nil {
		return fmt.Errorf("could not close connection: %w", err)
	}
	c.done <- true

	c.connection = nil
	c.done = nil
	c.enabled = false

	return nil
}

// RegisterWebhook implements plugin.Webhooker.
func (c *Plugin) RegisterWebhook(basePath string, g *gin.RouterGroup) {
}

// GetDisplay implements plugin.Displayer
// Invoked when the user views the plugin settings. Plugins do not need to be enabled to handle GetDisplay calls.
func (c *Plugin) GetDisplay(location *url.URL) string {
	if c.userCtx.Admin {
		if c.err != nil {
			return fmt.Sprintf("There has been an error: %v", c.err)
		}
		return fmt.Sprintf("This plugin requires a client token to be configured. Please see %s for more information", GetGotifyPluginInfo().ModulePath)
	} else {
		return "You are **NOT** an admin! You can do nothing:("
	}
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{userCtx: ctx}
}

// SetMessageHandler implements plugin.Messenger
// Invoked during initialization
func (c *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	c.msgHandler = h
}

func main() {
	panic("this should be built as go plugin")
}

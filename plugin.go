package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "https://github.com/david-kalmakoff/gotify-smtp-emailer",
		Version:     "0.2.3",
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

// ============================================================================

// Enable is called when the plugin is enabled
func (c *Plugin) Enable() error {
	if c.config == nil {
		if c.msgHandler != nil {
			c.msgHandler.SendMessage(plugin.Message{
				Title:   "SMTP Emailer: Error",
				Message: fmt.Sprintf("no config set"),
			})
		}
		return fmt.Errorf("no config set")
	}

	err := c.config.IsValid()
	if err != nil {
		if c.msgHandler != nil {
			c.msgHandler.SendMessage(plugin.Message{
				Title:   "SMTP Emailer: Error",
				Message: fmt.Sprintf("config is not valid: %v", err),
			})
		}
		return fmt.Errorf("config is invalid: %w", err)
	}

	// Start websocket connection
	go c.handleWS()

	c.enabled = true

	return nil
}

// ============================================================================

func (c *Plugin) handleWS() {
	var err error

	// start websocket connection
	c.connection, err = c.config.getWSConnection()
	if err != nil {
		if c.msgHandler != nil {
			c.msgHandler.SendMessage(plugin.Message{
				Title:   "SMTP Emailer: Error",
				Message: fmt.Sprintf("could not get ws connection: %v", err),
			})
		}
		c.Disable()
		log.Printf("SMTP Emailer: could not get ws connection: %v\n", err)
		return
	}

	if c.config.Environment == "development" && c.msgHandler != nil {
		c.msgHandler.SendMessage(plugin.Message{
			Title:   "SMTP Emailer: Enabled",
			Message: "Plugin has been enabled",
		})
	}

	c.done = make(chan bool)

	go c.handleMessages()

	// Send test message every 10 seconds in development environment
	if c.config.Environment == "development" && c.msgHandler != nil {
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
}

// ============================================================================

func (c *Plugin) handleMessages() {
	defer close(c.done)

	for {
		select {
		case <-c.done:
			return
		default:
			msg := plugin.Message{}

			// Read message from Gotify
			err := c.connection.ReadJSON(&msg)
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					return
				}
				log.Printf("SMTP Emailer: connection read error: %v\n", err)
				if c.config.Environment == "development" && c.msgHandler != nil {
					c.msgHandler.SendMessage(plugin.Message{
						Title:   "SMTP Emailer: Error",
						Message: fmt.Sprintf("could not read message: %v", err),
					})
				}
				continue
			}

			// Do not send email for internal messages
			if strings.Contains(msg.Title, "SMTP Emailer: ") {
				continue
			}

			// send message to smtp
			err = c.config.Smtp.Send(msg.Title, msg.Message)
			if err != nil {
				log.Printf("SMTP Emailer: smtp send error: %v\n", err)
				if c.msgHandler != nil {
					c.msgHandler.SendMessage(plugin.Message{
						Title:   "SMTP Emailer: Error",
						Message: fmt.Sprintf("smtp send error: %v", err),
					})
				}
				continue
			}

		}
	}
}

// ============================================================================

// getWSConnection is used to establish as websocket connection with Gotify
func (c *Config) getWSConnection() (*websocket.Conn, error) {
	count := 0
	for {
		count++
		uri := fmt.Sprintf("%s/stream?token=%s", c.Hostname, c.Token)
		ws, _, err := websocket.DefaultDialer.Dial(uri, nil)
		if err == nil {
			return ws, nil
		}
		if count > 60 {
			return nil, fmt.Errorf("Cannot connect to websocket %q: %w", uri, err)

		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ============================================================================

// Disable is called when the plugin is disabled
func (c *Plugin) Disable() error {
	err := c.connection.Close()
	if err != nil {
		if c.config.Environment == "development" && c.msgHandler != nil {
			c.msgHandler.SendMessage(plugin.Message{
				Title:   "SMTP Emailer: Disabled",
				Message: fmt.Sprintf("could not close ws connection: %v", err),
			})
		}
		log.Printf("SMTP Emailer: could not close ws connection: %v\n", err)
	}

	if c.config.Environment == "development" && c.msgHandler != nil {
		c.msgHandler.SendMessage(plugin.Message{
			Title:   "SMTP Emailer: Disabled",
			Message: "Plugin has been disabled",
		})
	}

	c.done <- true

	c.connection = nil
	c.done = nil
	c.enabled = false

	return nil
}

// ============================================================================

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

package dependencies

import (
	"medbratishka/internal/handler"
)

func (d *Dependencies) AuthHandler() handler.Handler {
	if d.authHandler == nil {
		d.authHandler = handler.NewAuthHandler(d.AuthService(), d.Logger())
	}
	return d.authHandler
}

func (d *Dependencies) BindingsHandler() handler.Handler {
	if d.bindingsHandler == nil {
		d.bindingsHandler = handler.NewBindingsHandler(d.AuthService(), d.BindingsService(), d.Logger())
	}
	return d.bindingsHandler
}

func (d *Dependencies) ChatHandler() handler.Handler {
	if d.chatHandler == nil {
		d.chatHandler = handler.NewChatHandler(d.AuthService(), d.ChatService(), d.Logger())
	}
	return d.chatHandler
}

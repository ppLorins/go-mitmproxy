package def

//type MidJourneyBaseHttpRequestContext struct {
//	Header string `json:"header"`
//	UTime  string `json:"utime"`
//}

//type MidJourneyMsgHttpRequestContext struct {
//	MessageID     string `json:"message_id"`
//	SessionID     string `json:"session_id"`
//	Type          int    `json:"type"` // fixed to be 3
//	Nonce         string `json:"nonce"`
//	GuildID       string `json:"guild_id"`
//	ChannelID     string `json:"channel_id"`
//	ApplicationID string `json:"application_id"`
//	Data          struct {
//		ComponentType int    `json:"component_type"` // fixed to be 2
//		CustomID      string `json:"custom_id"`
//	} `json:"data"`
//	MessageFlags int `json:"message_flags"` // fixed to be 0
//}

//type ImagineRequest struct {
//	Type          int    `json:"type"`
//	ApplicationID string `json:"application_id"`
//	GuildID       string `json:"guild_id"`
//	ChannelID     string `json:"channel_id"`
//	SessionID     string `json:"session_id"`
//	Data          struct {
//		Version string `json:"version"`
//		ID      string `json:"id"`
//		Name    string `json:"name"`
//		Type    int    `json:"type"`
//		Options []struct {
//			Type  int    `json:"type"`
//			Name  string `json:"name"`
//			Value string `json:"value"`
//		} `json:"options"`
//		ApplicationCommand struct {
//			ID                       string      `json:"id"`
//			ApplicationID            string      `json:"application_id"`
//			Version                  string      `json:"version"`
//			DefaultMemberPermissions interface{} `json:"default_member_permissions"`
//			Type                     int         `json:"type"`
//			Nsfw                     bool        `json:"nsfw"`
//			Name                     string      `json:"name"`
//			Description              string      `json:"description"`
//			DmPermission             bool        `json:"dm_permission"`
//			Contexts                 interface{} `json:"contexts"`
//			Options                  []struct {
//				Type        int    `json:"type"`
//				Name        string `json:"name"`
//				Description string `json:"description"`
//				Required    bool   `json:"required"`
//			} `json:"options"`
//		} `json:"application_command"`
//		Attachments []interface{} `json:"attachments"`
//	} `json:"data"`
//	Nonce string `json:"nonce"`
//}

package constant

const (
	TelegramMessageTemplate = "<b>🚀 Database Backup Report</b>\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"<b>✅ Status:</b> <code>Success</code>\n" +
		"<b>🗄️ Database:</b> <code>%s</code>\n" +
		"<b>📂 File:</b> <code>%s</code>\n" +
		"<b>⏰ Finished:</b> <code>%s</code>\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"<i>#GoBackup #AutomatedNotification</i>"

	TelegramErrorTemplate = "<b>⚠️ Backup FAILED!</b>\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"<b>❌ Status:</b> <code>Error</code>\n" +
		"<b>🗄️ Database:</b> <code>%s</code>\n" +
		"<b>🔴 Message:</b> <code>%s</code>\n" +
		"<b>⏰ Time:</b> <code>%s</code>\n" +
		"━━━━━━━━━━━━━━━━━━━━━━\n" +
		"<i>#GoBackup #AlertSystem</i>"
)

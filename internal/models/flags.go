package models

type PersistentFlags struct {
	Output     *string
	LogLevel   *string
	ShowCaller *bool
}

type CommandFlags struct {
	Name    *string
	Source  *string
	Mode    *string
	Enabled *bool
}

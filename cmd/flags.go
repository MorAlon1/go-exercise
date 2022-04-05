package cmd

//consts
var (
	pathFlagName            = "path"
	excludeFlagName         = "exclude"
	destinationRelativePath = "/output"
)

// Scan flags
var (
	// repositoryUrl the path to the repository to scan
	path    string
	exclude string
)

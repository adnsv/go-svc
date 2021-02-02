package svcapp

func servicePath(name string) string {
	// assuming system service, not user service
	return "/Library/LaunchDaemons/" + name + ".plist"
}

// tbd: implement

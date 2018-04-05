package bx

func GetBluemixZones() []string {
	return []string{
		"ap-north",
		"jp-tok",
		"ap-south",
		"au-syd",
		"eu-central",
		"eu-de",
		"uk-south",
		"eu-gb",
		"us-east",
		"us-south",
	}
}

func GetBluemixLocations() []string {
	return []string{
		"ams03",
		"fra02",
		"par01",
	}
}

func GetBluemixMachineTypes() []string {
	return []string{
		"b2c.16x64",
		"b2c.32x128",
		"b2c.4x16",
		"b2c.56x242",
		"dal10",
		"dal12",
		"dal13",
		"mb1c.16x64",
		"mb1c.4x32",
		"md1c.16x64.4x4tb",
		"md1c.28x512.4x4tb",
		"mr1c.28x512",
		"sao01",
		"u2c.2x4",
	}
}

func HardwareIsolationLevels() []string {
	return []string{
		"dedicated",
		"shared",
	}
}

package banner

import (
	"fmt"
	"strings"
)

// FlowBanner 是简洁版的Flow ASCII艺术
const FlowBanner = `
  _____  _                
 |  ___|| |  ___  __      __
 | |_   | | / _ \ \ \ /\ / /
 |  _|  | || (_) | \ V  V / 
 |_|    |_| \___/   \_/\_/  %s
                      %s
`

// SmallFlowBanner 是更小的Flow ASCII艺术
const SmallFlowBanner = `
  __ _                
 / _| | _____      __
| |_| |/ _ \ \ /\ / /
|  _| | (_) \ V  V / 
|_| |_|\___/ \_/\_/  %s
              %s
`

// MicroFlowBanner 是极简的Flow ASCII艺术
const MicroFlowBanner = `
 ___ _     ___  _    _ 
| __| |___| _ \| |  | |
| _|| / _ \  _/| |__| |
|_| |_\___/_|  |____|_|  %s
                  %s
`

// Print 打印Flow标志
func Print(version, description string) {
	fmt.Printf(SmallFlowBanner, version, description)
}

// PrintWithSize 打印指定大小的Flow标志
// size可以是："micro", "small", "normal"
func PrintWithSize(version, description, size string) {
	var banner string

	switch strings.ToLower(size) {
	case "micro":
		banner = MicroFlowBanner
	case "normal":
		banner = FlowBanner
	default:
		banner = SmallFlowBanner
	}

	fmt.Printf(banner, version, description)
}

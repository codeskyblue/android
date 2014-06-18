package adb

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type DeviceType int
type SdkVersion int

const (
	PHONE DeviceType = iota
	TABLET_7
	TABLET_10
)

const (
	ECLAIR SdkVersion = iota + 7
	FROYO
	GINGERBREAD
	GINGERBREAD_MR1
	HONEYCOMB
	HONEYCOMB_MR1
	HONEYCOMB_MR2
	ICE_CREAM_SANDWICH
	ICE_CREAM_SANDWICH_MR1
	JELLY_BEAN
	JELLY_BEAN_MR1
	JELLY_BEAN_MR2
	KITKAT
	LATEST = KITKAT
)

var sdkMap = map[SdkVersion]string{
	ECLAIR:                 `ECLAIR`,
	FROYO:                  `FROYO`,
	GINGERBREAD:            `GINGERBREAD`,
	GINGERBREAD_MR1:        `GINGERBREAD_MR1`,
	HONEYCOMB:              `HONEYCOMB`,
	HONEYCOMB_MR1:          `HONEYCOMB_MR1`,
	HONEYCOMB_MR2:          `HONEYCOMB_MR2`,
	ICE_CREAM_SANDWICH:     `ICE_CREAM_SANDWICH`,
	ICE_CREAM_SANDWICH_MR1: `ICE_CREAM_SANDWICH_MR1`,
	JELLY_BEAN:             `JELLY_BEAN`,
	JELLY_BEAN_MR1:         `JELLY_BEAN_MR1`,
	JELLY_BEAN_MR2:         `JELLY_BEAN_MR2`,
	KITKAT:                 `KITKAT`,
}

type Device struct {
	Dialer       `json:"-"`
	Serial       string     `json:"serial"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Sdk          SdkVersion `json:"sdk"`
	Version      string     `json:"version"`
	Density      int64      `json:"density"`
}

type DeviceFilter struct {
	Type    DeviceType
	Serials []string
	Density int64
	MinSdk  SdkVersion
	MaxSdk  SdkVersion
}

var (
	AllDevices = &DeviceFilter{MaxSdk: LATEST}
)

func (s SdkVersion) String() string {
	return sdkMap[s]
}

// filter -f "serials=[...];type=tablet;count=5;version >= 4.1.1;"

/*func GetFilter(arg string) {*/

/*}*/

type DeviceWatcher []chan []*Device

func (d *Device) Transport(conn *AdbConn) error {
	return conn.TransportSerial(d.Serial)
}

func (adb *Adb) ParseDevices(filter *DeviceFilter, input []byte) []*Device {
	lines := strings.Split(string(input), "\n")

	devices := make([]*Device, 0, len(lines))

	var wg sync.WaitGroup

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			device := strings.Split(line, "\t")[0]

			d := &Device{Dialer: adb.Dialer, Serial: device}
			devices = append(devices, d)

			wg.Add(1)
			go func() {
				defer wg.Done()
				d.Update()
			}()
		}
	}

	wg.Wait()

	result := make([]*Device, 0, len(lines))
	for _, device := range devices {
		if device.MatchFilter(filter) {
			result = append(result, device)
		}
	}
	return result
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return len(list) == 0
}

func (d *Device) MatchFilter(filter *DeviceFilter) bool {
	if filter == nil {
		return true
	}

	if d.Sdk < filter.MinSdk {
		return false
	} else if filter.MaxSdk != 0 && d.Sdk > filter.MaxSdk {
		fmt.Println(d.Sdk)
		return false
	} else if !stringInSlice(d.Serial, filter.Serials) {
		return false
	} else if filter.Density != 0 && filter.Density != d.Density {
		return false
	}
	return true
}

func (d *Device) GetProp(prop string) chan string {
	out := make(chan string)
	go func() {
		p := ShellSync(d, "getprop", prop)
		out <- strings.TrimSpace(string(p))
	}()

	return out
}

func (d *Device) SetScreenOn(on bool) {
	current := d.findValue("mScreenOn=false", "dumpsys", "input_method")
	if !current && on || current && !on {
		d.SendKey(26)
	}
}

func (d *Device) findValue(val string, args ...string) bool {
	out := Shell(d, args...)
	current := false
	for line := range out {
		if line != nil {
			current = bytes.Contains(line, []byte(val))
		}
	}
	return current
}

func (d *Device) SendKey(aKey int) {
	ShellSync(d, "input", "keyevent", fmt.Sprintf("%d", aKey))
}

func (d *Device) Unlock() {
	current := d.findValue("mLockScreenShown true", "dumpsys", "activity")
	if current {
		d.SendKey(82)
	}
}

func (d *Device) Update() {

	out := []chan string{
		d.GetProp("ro.product.manufacturer"),
		d.GetProp("ro.product.model"),
		d.GetProp("ro.build.version.release"),
		d.GetProp("ro.build.version.sdk"),
		d.GetProp("ro.sf.lcd_density"),
	}

	d.Manufacturer = <-out[0]
	d.Model = <-out[1]
	d.Version = <-out[2]
	sdk := <-out[3]
	den := <-out[4]

	sdk_int, _ := strconv.ParseInt(sdk, 10, 0)
	d.Sdk = SdkVersion(sdk_int)

	d.Density, _ = strconv.ParseInt(den, 10, 0)
}

func (d *Device) String() string {
	return fmt.Sprintf("%s\t%s %s\t[%s (%s)]", d.Serial, d.Manufacturer, d.Model, d.Version, sdkMap[d.Sdk])
}

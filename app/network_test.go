// SPDX-License-Identifier: Unlicense OR MIT

package app

import "testing"

func TestNetStateMoves(t *testing.T) {
	var s netState
	check := func(name string, info NetworkInfo, news bool, wantNews, wantMoved bool) {
		t.Helper()
		if news != wantNews || info.Moved != wantMoved {
			t.Errorf("%s: news %v moved %v, want news %v moved %v", name, news, info.Moved, wantNews, wantMoved)
		}
	}
	info, news := s.observe("100", "wlan0", []string{"192.168.1.14"}, 0, false)
	check("first network", info, news, true, false)
	if !info.MeteredKnown || info.Metered {
		t.Errorf("first report carries the metered flag: %+v", info)
	}
	info, news = s.observe("100", "wlan0", []string{"192.168.1.14"}, 0, false)
	check("repeat", info, news, false, false)
	info, news = s.observe("100", "wlan0", []string{"192.168.1.14", "2001:db8::5"}, 0, false)
	check("address added", info, news, false, false)
	info, news = s.observe("100", "wlan0", []string{"2001:db8::5"}, 0, false)
	check("address gone", info, news, true, true)
	info, news = s.observe("200", "tun0", nil, 1, true)
	check("a VPN is not a network", info, news, false, false)
	info, news = s.observe("201", "rmnet_data0", []string{"10.64.0.2"}, 1, false)
	check("to mobile data", info, news, true, true)
	if !info.Metered || info.Interface != "rmnet_data0" {
		t.Errorf("mobile data is metered on rmnet_data0: %+v", info)
	}
	info, news = s.observe("201", "rmnet_data0", []string{"10.64.0.2"}, 0, false)
	check("metered flip only", info, news, true, false)
	info, news = s.observe("201", "rmnet_data0", []string{"10.64.0.2"}, -1, false)
	check("unknown metered is not news", info, news, false, false)
	s.lost("201")
	info, news = s.observe("201", "rmnet_data0", []string{"10.64.0.2"}, 0, false)
	check("back after loss", info, news, true, true)
}

func TestNetStateUnknownMeteredFirst(t *testing.T) {
	var s netState
	info, news := s.observe("1", "eth0", []string{"10.0.0.2"}, -1, false)
	if !news || info.MeteredKnown {
		t.Fatalf("first report is news and says the flag is unknown: %+v %v", info, news)
	}
}

func TestWatchNetworkDelivers(t *testing.T) {
	var got []NetworkInfo
	stop, err := WatchNetwork(func(i NetworkInfo) { got = append(got, i) })
	if err != nil {
		t.Fatal(err)
	}
	networkChanged("1", "eth0", []string{"10.0.0.2"}, 0, false)
	networkChanged("2", "wlan0", []string{"10.0.1.2"}, 1, false)
	stop()
	networkChanged("3", "wlan1", []string{"10.0.2.2"}, 1, false)
	if len(got) != 2 || !got[1].Moved || !got[1].Metered {
		t.Fatalf("got %+v", got)
	}
}

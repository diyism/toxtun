package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	tox "github.com/TokTok/go-toxcore-c"
	"github.com/bitly/go-simplejson"
)

var servers = []interface{}{
	"85.172.30.117", uint16(33445), "8E7D0B859922EF569298B4D261A8CCB5FEA14FB91ED412A7603A585A25698832",
	"205.185.116.116", uint16(33445), "A179B09749AC826FF01F37A9613F6B57118AE014D4196A0E1105A98F93A54702",
	"biribiri.org", uint16(33445), "F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67",
	"80.87.193.193", uint16(33445), "B38255EE4B054924F6D79A5E6E5889EC94B6ADF6FE9906F97A3D01E3D083223A",
	"51.254.84.212", uint16(33445), "AEC204B9A4501412D5F0BB67D9C81B5DB3EE6ADA64122D32A3E9B093D544327D",
	// "127.0.0.1", uint16(33445), "398C8161D038FD328A573FFAA0F5FAAF7FFDE5E8B4350E7D15E6AFD0B993FC52",
}

var tox_savedata_fname string
var tox_disable_udp = false

func init() {
	flag.BoolVar(&tox_disable_udp, "disable-udp", tox_disable_udp,
		fmt.Sprintf("if tox disable udp, default: %v", tox_disable_udp))
	flag.BoolVar(&useFixedBSs, "use-fixedbs", useFixedBSs, "use fixed bootstraps, possible faster.")
}

func makeTox(name string) *tox.Tox {
	tox_savedata_fname = fmt.Sprintf("./%s.data", name)
	var nickPrefix = fmt.Sprintf("%s.", name)
	var statusText = fmt.Sprintf("%s of toxtun", name)

	opt := tox.NewToxOptions()
	opt.Udp_enabled = !tox_disable_udp

	if tox.FileExist(tox_savedata_fname) {
		data, err := ioutil.ReadFile(tox_savedata_fname)
		if err != nil {
			errl.Println(err)
		} else {
			opt.Savedata_data = data
			opt.Savedata_type = tox.SAVEDATA_TYPE_TOX_SAVE
		}
	}
	port := 33445
	var t *tox.Tox
	for i := 0; i < 71; i++ {
		opt.Tcp_port = uint16(port + i)
		// opt.Tcp_port = 0
		t = tox.NewTox(opt)
		if t != nil {
			break
		}
	}
	if t == nil {
		panic(nil)
	}
	if tox_disable_udp {
		info.Println("TCP port:", opt.Tcp_port)
	} else {
		info.Println("TCP port:", "disabled")
	}
	if false {
		time.Sleep(1 * time.Hour)
	}

	for i := 0; i < len(servers)/3; i++ {
		if useFixedBSs {
			// continue
		}
		r := i * 3
		ipstr, port, pubkey := servers[r+0].(string), servers[r+1].(uint16), servers[r+2].(string)
		r1, err := t.Bootstrap(ipstr, port, pubkey)
		r2, err := t.AddTcpRelay(ipstr, port, pubkey)
		info.Println("bootstrap:", r1, err, r2, i, r, ipstr, port)
	}
	if useFixedBSs {
		addFixedBootstraps(t)
	}

	pubkey := t.SelfGetPublicKey()
	seckey := t.SelfGetSecretKey()
	toxid := t.SelfGetAddress()
	debug.Println("keys:", pubkey, seckey, len(pubkey), len(seckey))
	info.Println("toxid:", toxid)

	defaultName := t.SelfGetName()
	humanName := nickPrefix + toxid[0:5]
	if humanName != defaultName {
		t.SelfSetName(humanName)
	}
	humanName = t.SelfGetName()
	debug.Println(humanName, defaultName)

	defaultStatusText, err := t.SelfGetStatusMessage()
	if defaultStatusText != statusText {
		t.SelfSetStatusMessage(statusText)
	}
	debug.Println(statusText, defaultStatusText, err)

	sz := t.GetSavedataSize()
	sd := t.GetSavedata()
	debug.Println("savedata:", sz, t)
	debug.Println("savedata", len(sd), t)

	err = t.WriteSavedata(tox_savedata_fname)
	debug.Println("savedata write:", err)

	// add friend norequest
	fv := t.SelfGetFriendList()
	for _, fno := range fv {
		fid, err := t.FriendGetPublicKey(fno)
		if err != nil {
			debug.Println(err)
		} else {
			t.FriendAddNorequest(fid)
		}
	}
	debug.Println("added friends:", len(fv))

	return t
}

func iterate(t *tox.Tox) {
	// toxcore loops
	shutdown := false
	loopc := 0
	itval := 0
	if !shutdown {
		iv := t.IterationInterval()
		if iv != itval {
			if itval-iv > 20 || iv-itval > 20 {
				// debug.Println("tox itval changed:", itval, iv)
			}
			itval = iv
		}

		t.Iterate()
		status := t.SelfGetConnectionStatus()
		if loopc%5500 == 0 {
			if status == 0 {
				// debug.Printf(".")
			} else {
				// debug.Printf("%d,", status)
			}
		}
		loopc += 1
		// time.Sleep(50 * time.Millisecond)
	}

	// t.Kill()
}

var useFixedBSs = false

func addFixedBootstraps(t *tox.Tox) {
	if useFixedBSs {
		{
			node := []interface{}{"cotox.tk", uint16(33445), "AF66C5FFAA6CA67FB8E287A5B1D8581C15B446E12BF330963EF29E3AFB692918"}
			_, err := t.Bootstrap(node[0].(string), node[1].(uint16), node[2].(string))
			_, err = t.AddTcpRelay(node[0].(string), node[1].(uint16), node[2].(string))
			log.Println("hehehe", err == nil)
		}
		if false {
			node := []interface{}{"133.130.127.155", uint16(33445), "5EE85FD7B4B6BD8FD113A1E8CC5853A233008B574E07F2CC76A7EA43AE24AE07"}
			_, err := t.Bootstrap(node[0].(string), node[1].(uint16), node[2].(string))
			//_, err = t.AddTcpRelay(node[0].(string), node[1].(uint16), node[2].(string))
			log.Println("hehehe", err == nil)
		}
	}
}

// 切换到其他的bootstrap nodes上
func switchServer(t *tox.Tox) {
	newNodes := get3nodes()
	for _, node := range newNodes {
		if useFixedBSs {
			// continue
		}
		r1, err := t.Bootstrap(node.ipaddr, node.port, node.pubkey)
		if node.status_tcp {
			r2, err := t.AddTcpRelay(node.ipaddr, node.port, node.pubkey)
			info.Println("bootstrap(tcp):", r1, err, r2, node.ipaddr, node.last_ping, node.status_tcp)
		} else {
			info.Println("bootstrap(udp):", r1, err, node.ipaddr,
				node.last_ping, node.status_tcp, node.last_ping_rt)
		}
	}
	currNodes = newNodes

	addFixedBootstraps(t)
}

func get3nodes() (nodes [3]ToxNode) {
	idxes := make(map[int]bool, 0)
	currips := make(map[string]bool, 0)
	for idx := 0; idx < len(currNodes); idx++ {
		currips[currNodes[idx].ipaddr] = true
	}
	for n := 0; n < len(allNodes)*3; n++ {
		idx := rand.Int() % len(allNodes)
		_, ok1 := idxes[idx]
		_, ok2 := currips[allNodes[idx].ipaddr]
		if !ok1 && !ok2 && allNodes[idx].status_tcp == true && allNodes[idx].last_ping_rt > 0 {
			idxes[idx] = true
			if len(idxes) == 3 {
				break
			}
		}
	}
	if len(idxes) < 3 {
		errl.Println("can not find 3 new nodes:", idxes)
	}

	_idx := 0
	for k, _ := range idxes {
		nodes[_idx] = allNodes[k]
		_idx += 1
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
	initThirdPartyNodes()
	initToxNodes()
	// go pingNodes()
}

// fixme: chown root.root toxtun-go && chmod u+s toxtun-go
// should block
func pingNodes() {
	stop := false
	for !stop {
		btime := time.Now()
		errcnt := 0
		for idx, node := range allNodes {
			if false {
				log.Println(idx, node)
			}
			if true {
				// rtt, err := Ping0(node.ipaddr, 3)
				rtt, err := Ping0(node.ipaddr, 3)
				if err != nil {
					// log.Println("ping", ok, node.ipaddr, rtt.String())
					log.Println("ping", err, node.ipaddr, rtt.String())
					errcnt += 1
				}
				if err == nil {
					allNodes[idx].last_ping_rt = uint(time.Now().Unix())
					allNodes[idx].rtt = rtt
				} else {
					allNodes[idx].last_ping_rt = uint(0)
					allNodes[idx].rtt = time.Duration(0)
				}
			}
		}
		etime := time.Now()
		log.Printf("Pinged all=%d, errcnt=%d, %v\n", len(allNodes), errcnt, etime.Sub(btime))

		// TODO longer ping interval
		time.Sleep(30 * time.Second)
	}
}

func initThirdPartyNodes() {
	for idx := 0; idx < 3*3; idx += 3 {
		node := ToxNode{
			isthird:      true,
			ipaddr:       servers[idx].(string),
			port:         servers[idx+1].(uint16),
			pubkey:       servers[idx+2].(string),
			last_ping:    uint(time.Now().Unix()),
			last_ping_rt: uint(time.Now().Unix()),
			status_tcp:   true,
		}

		allNodes = append(allNodes, node)
	}
}

func initToxNodes() {
	bcc, err := Asset("toxnodes.json")
	if err != nil {
		log.Panicln(err)
	}
	jso, err := simplejson.NewJson(bcc)
	if err != nil {
		log.Panicln(err)
	}

	nodes := jso.Get("nodes").MustArray()
	for idx := 0; idx < len(nodes); idx++ {
		nodej := jso.Get("nodes").GetIndex(idx)
		/*
			log.Println(idx, nodej.Get("ipv4"), nodej.Get("port"), nodej.Get("last_ping"),
				len(nodej.Get("tcp_ports").MustArray()))
		*/
		node := ToxNode{
			ipaddr:       nodej.Get("ipv4").MustString(),
			port:         uint16(nodej.Get("port").MustUint64()),
			pubkey:       nodej.Get("public_key").MustString(),
			last_ping:    uint(nodej.Get("last_ping").MustUint64()),
			status_tcp:   nodej.Get("status_tcp").MustBool(),
			last_ping_rt: uint(time.Now().Unix()),
			weight:       calcNodeWeight(nodej),
		}

		allNodes = append(allNodes, node)
		if idx < len(currNodes) {
			currNodes[idx] = node
		}
	}

	sort.Sort(ByRand(allNodes))
	for idx, node := range allNodes {
		if false {
			log.Println(idx, node.ipaddr, node.port, node.last_ping)
		}
	}
	info.Println("Load nodes:", len(allNodes))
}

func calcNodeWeight(nodej *simplejson.Json) int {
	return 0
}

var allNodes = make([]ToxNode, 0)
var currNodes [3]ToxNode

type ToxNode struct {
	isthird    bool
	ipaddr     string
	port       uint16
	pubkey     string
	weight     int
	usetimes   int
	legacy     int
	chktimes   int
	last_ping  uint
	status_tcp bool
	///
	last_ping_rt uint // 程序内ping的时间
	rtt          time.Duration
}

type ByRand []ToxNode

func (this ByRand) Len() int           { return len(this) }
func (this ByRand) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this ByRand) Less(i, j int) bool { return rand.Int()%2 == 0 }

var livebots = []string{
	"56A1ADE4B65B86BCD51CC73E2CD4E542179F47959FE3E0E21B4B0ACDADE51855D34D34D37CB5", // groupbot
	"76518406F6A9F2217E8DC487CC783C25CC16A15EB36FF32E335A235342C48A39218F515C39A6", //echobot@toxme.io
	"DD7A68B345E0AA918F3544AA916B5CA6AED6DE80389BFF1EF7342DACD597943D62BDEED1FC67", // my echobot
	"03F47F0AE26BE32C73579CBA2C5421A159EDFF74535A7E8C6480398D93A0EA2E02B1B20B80D7", // DobroBot
	"A922A51E1C91205B9F7992E2273107D47C72E8AE909C61C28A77A4A2A115431B14592AB38A3B", // toxirc
	"5EE85FD7B4B6BD8FD113A1E8CC5853A233008B574E07F2CC76A7EA43AE24AE0754DBD6B8FD3F", // ToxIRCBotCN
	"415732B8A549B2A1F9A278B91C649B9E30F07330E8818246375D19E52F927C57F08A44E082F6", // LainBot
	"398C8161D038FD328A573FFAA0F5FAAF7FFDE5E8B4350E7D15E6AFD0B993FC529FA90C343627", // envoy
}

func addLiveBots(t *tox.Tox) {
	for _, botid := range livebots {
		t.FriendAdd(botid, "hello")
	}
}

func livebotsOnFriendConnectionStatus(t *tox.Tox, friendNumber uint32, status int) {
	fid, _ := t.FriendGetPublicKey(friendNumber)
	if strings.HasPrefix(livebots[5], fid) {
		t.FriendSendMessage(friendNumber, "/mute on")
	}
}

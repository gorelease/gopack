package main

import (
	"encoding/json"
	"fmt"
	"flag"
	"image"
	_ "image/color"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

type serverJSON struct {
	Apiserver, Imgserver, Fullbg, Imgurl, Challenge string
}

var (
	vm = otto.New()
)

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	vm.Run(gtjsFunctions)

	//265g 482708a0f54c0bff54c6dafe3629d7bc
	//douban 2add92d43d3beb28149720a6a3698061
	//geetest a40fd3b0d712165c5d13e6f747e948d4
	pCount := flag.Int("c", 100, "count")
	pID := flag.String("gt", "482708a0f54c0bff54c6dafe3629d7bc", "geetest ID")
	flag.Parse()

	for i := 0; i < *pCount; i++ {
		fmt.Println(flirt(pID))
	}
}

func flirt(gtID *string) string {
	challengePage, err := http.Get(challengeURL + *gtID)
	if err != nil {
		log.Fatal(err)
	}
	defer challengePage.Body.Close()
	challengeBody, _ := ioutil.ReadAll(challengePage.Body)
	challengeString := string(challengeBody)
	jsonString := challengeString[strings.Index(challengeString, "= {") + 2 : strings.Index(challengeString, "}") + 1]
	config := serverJSON{}
	json.Unmarshal([]byte(jsonString), &config)
	img1r, err := http.Get(config.Imgserver + config.Fullbg)
	if err != nil {
		log.Fatal(err)
	}
	defer img1r.Body.Close()
	img2r, err := http.Get(config.Imgserver + config.Imgurl)
	if err != nil {
		log.Fatal(err)
	}
	defer img1r.Body.Close()

	image1, _, _ := image.Decode(img1r.Body)
	image2, _, _ := image.Decode(img2r.Body)

	xPos := deCAPTCHA(image1, image2)
	moveTrack := generateMoveTrack(xPos)
	passTime := moveTrack[len(moveTrack) - 1][2]

	vm.Set("apiserver", config.Apiserver)
	vm.Set("challenge", config.Challenge)
	vm.Set("xr", xPos)
	vm.Set("b", passTime)
	vm.Set("a", moveTrack)

	jsResult, _ := vm.Run(gtjsGetResponseURL)
	ajaxURL, _ := jsResult.ToString()
	ajaxPage, err := http.Get(ajaxURL)
	if err != nil {
		log.Fatal(err)
	}
	defer ajaxPage.Body.Close()
	ajaxPageString, _ := ioutil.ReadAll(ajaxPage.Body)
	return string(ajaxPageString)
}

func deCAPTCHA(image1, image2 image.Image) int {
	for x := image1.Bounds().Min.X; x < image1.Bounds().Max.X; x++ {
		for y := image1.Bounds().Min.Y; y < image1.Bounds().Max.Y; y++ {
			fr, fb, fg, _ := image1.At(x, y).RGBA()
			br, bb, bg, _ := image2.At(x, y).RGBA()
			if math.Abs(float64(fr) - float64(br)) > 5000.0 && math.Abs(float64(fb) - float64(bb)) > 5000.0 && math.Abs(float64(fg) - float64(bg)) > 5000.0 {
				return x - 3
			}
		}
	}
	log.Fatal("rare! deCAPTACH failed")
	return 0
}

func generateMoveTrack(xPosAnswer int) [][]int {
	totalFrames := int(xPosAnswer / 2) + rand.Intn(5)

	moveTrack := make([][]int, totalFrames)
	moveTrack[0] = []int{int(-rand.NormFloat64() * 8.0 - 20.0), int(-rand.NormFloat64() * 8.0 - 20.0), 0}
	moveTrack[1] = []int{0, 0, 0}

	periodParam := rand.Float64()
	for i := 2; i < totalFrames; i++ {
		moveTrack[i] = []int{0, int(math.Sin(float64(i) * periodParam * 0.08) * 4.0), i * 8 + rand.Intn(5)}
	}

	xPosBitmap := make([]bool, xPosAnswer)
	for i := 0; i < totalFrames - 4; {
		t := &xPosBitmap[rand.Intn(xPosAnswer - 1)];
		if (!*t) {
			*t = true
			i++
		}
	}
	xPosBitmap[xPosAnswer - 1] = true;

	k := 2
	for i, v := range xPosBitmap {
		if (v) {
			moveTrack[k][0] = i + 1
			k++
		}
	}

	copy(moveTrack[totalFrames - 1], moveTrack[totalFrames - 2])
	moveTrack[totalFrames - 1][2] += 100 + rand.Intn(300)

	return moveTrack
}

const (
	challengeURL = "http://api.geetest.com/get.php?gt="

	// copy from geetest.2.8.4.js
	gtjsFunctions = `
		function p() {
			for (var b = challenge.slice(32), c = [], d = 0; d < b.length; d++) {
				var e = b.charCodeAt(d);
				c[d] = e > 57 ? e - 87 : e - 48
			}
			b = 36 * c[0] + c[1];
			var f, g = Math.round(xr) + b, h = challenge.slice(0, 32), j = [[], [], [], [], []], k = {}, l = 0;
			d = 0;
			for (var m = h.length; m > d; d++)f = h.charAt(d), k[f] || (k[f] = 1, j[l].push(f), l++, l = 5 == l ? 0 : l);
			for (var n, o = g, p = 4, q = "", r = [1, 2, 5, 10, 50]; o > 0;)o - r[p] >= 0 ? (n = parseInt(Math.random() * j[p].length, 10), q += j[p][n], o -= r[p]) : (j.splice(p, 1), r.splice(p, 1), p -= 1);
			return q
		}

		function n() {
			var a = parseInt(1e6 * Math.random()), b = 1e6 - a;
			return a + "." + b
		}

		function z() {
			var b = function (a) {
				for (var b = [], c = 0; c < a.length - 1; c++) {
					var d = [];
					d[0] = a[c + 1][0] - a[c][0], d[1] = a[c + 1][1] - a[c][1], d[2] = a[c + 1][2] - a[c][2], (0 !== d[0] || 0 !== d[1] || 0 !== d[2]) && b.push(d)
				}
				return b
			}, c = function (a) {
				var b = "!$'()*+,-./0123456789:;?@ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~";
				return b.charAt(a)
			}, d = function (a) {
				for (var b = "", d = parseInt(a); d;)b = b.toString() + c(d % 77 + 2), d = (d - d % 77) / 77;
				return b ? b : c(2)
			}, e = function (a) {
				for (var e = b(a), f = "", g = "", h = "", i = 0; i < e.length; i++) {
					var j = "", k = "", l = "", m = parseInt(e[i][0]), n = parseInt(e[i][1]), o = parseInt(e[i][2]);
					-2 == n && m >= 1 && 4 >= m ? k = c(1 + m) : -1 == n && m >= -2 && 6 >= m ? k = c(8 + m) : 0 === n && m >= -5 && 10 >= m ? k = c(20 + m) : 1 == n && m >= -2 && 7 >= m ? k = c(33 + m) : (n >= -17 && 20 >= n ? k = c(58 + n) : -17 > n ? k = c(1) + d(-18 - n) + c(0) : n > 20 && (k = c(0) + d(n - 21) + c(1)), m >= -21 && 55 >= m ? j = c(23 + m) : -21 > m ? j = c(1) + d(-22 - m) + c(0) : m > 55 && (j = c(0) + d(m - 56) + c(1)));
					var p = d(o);
					l = p.length <= 1 ? p : 2 === p.length ? c(0) + p : c(0) + c(0), f += j, g += k, h += l
				}
				return f + c(1) + c(1) + c(1) + g + c(1) + c(1) + c(1) + h
			};
			return e(a)
		}
	`

	// copy from geetest.2.8.4.js
	gtjsGetResponseURL = `
		apiserver + "ajax.php?api=jordan&challenge=" + challenge + "&userresponse=" + p() + "&passtime=" + b + "&b=" + n() + "&imgload=" + parseInt(Math.random() * 200 + 300) + "&random=" + parseInt(1e6 * Math.random()) + "&a=" + z()
	`
)

package client_phone

// Uplink:   even number
// Downlink: odd number

var device_to_port = map[string][2]int {
	"xm00": {3230, 3231},
	"xm01": {3232, 3233},
	"xm02": {3234, 3235},
	"xm03": {3236, 3237},
	"xm04": {3238, 3239},
	"xm05": {3240, 3241},
	"xm06": {3242, 3243},
	"xm07": {3244, 3245},
	"xm08": {3246, 3247},
	"xm09": {3248, 3249},
	"xm10": {3250, 3251},
	"xm11": {3252, 3253},
	"xm12": {3254, 3255},
	"xm13": {3256, 3257},
	"xm14": {3258, 3259},
	"xm15": {3260, 3261},
	"xm16": {3262, 3263},
	"xm17": {3264, 3265},
	"sm00": {3200, 3201},
	"sm01": {3202, 3203},
	"sm02": {3204, 3205},
	"sm03": {3206, 3207},
	"sm04": {3208, 3209},
	"sm05": {3210, 3211},
	"sm06": {3212, 3213},
	"sm07": {3214, 3215},
	"sm08": {3216, 3217},
	"sm09": {3218, 3219},
	"qc00": {3270, 3271},
	"qc01": {3272, 3273},
	"qc02": {3274, 3275},
	"qc03": {3276, 3277},
	"unam": {3280, 3281},
}

var port_to_device = map[string]string {
	"3230": "xm00",
	"3231": "xm00",
	"3232": "xm01",
	"3233": "xm01",
	"3234": "xm02",
	"3235": "xm02",
	"3236": "xm03",
	"3237": "xm03",
	"3238": "xm04",
	"3239": "xm04",
	"3240": "xm05",
	"3241": "xm05",
	"3242": "xm06",
	"3243": "xm06",
	"3244": "xm07",
	"3245": "xm07",
	"3246": "xm08",
	"3247": "xm08",
	"3248": "xm09",
	"3249": "xm09",
	"3250": "xm10",
	"3251": "xm10",
	"3252": "xm11",
	"3253": "xm11",
	"3254": "xm12",
	"3255": "xm12",
	"3256": "xm13",
	"3257": "xm13",
	"3258": "xm14",
	"3259": "xm14",
	"3260": "xm15",
	"3261": "xm15",
	"3262": "xm16",
	"3263": "xm16",
	"3264": "xm17",
	"3265": "xm17",
	"3200": "sm00",
	"3201": "sm00",
	"3202": "sm01",
	"3203": "sm01",
	"3204": "sm02",
	"3205": "sm02",
	"3206": "sm03",
	"3207": "sm03",
	"3208": "sm04",
	"3209": "sm04",
	"3210": "sm05",
	"3211": "sm05",
	"3212": "sm06",
	"3213": "sm06",
	"3214": "sm07",
	"3215": "sm07",
	"3216": "sm08",
	"3217": "sm08",
	"3218": "sm09",
	"3219": "sm09",
	"3270": "qc00",
	"3271": "qc00",
	"3272": "qc01",
	"3273": "qc01",
	"3274": "qc02",
	"3275": "qc02",
	"3276": "qc03",
	"3277": "qc03",
	"3280": "unam",
	"3281": "unam",
}

// func main() {
// 	print(device_to_port)
// 	print(port_to_device)
// }
package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/kellydunn/golang-geo"
	"github.com/samber/lo"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
	"voice-dispatch/app/model"
	"voice-dispatch/app/request"
	"voice-dispatch/app/response"
	"voice-dispatch/config"
	"voice-dispatch/infra"
)

const DISPATCHKEY = "voice-dispatch"

var platformNum int
var startWorkSpaceNames = [4]string{"1200平台", "1230平台", "1250平台", "1280平台"}
var endWorkSpaceNames = [4]string{"1号排土场", "2号排土场", "1号破碎厂", "2号破碎厂"}
var startWorkSpaceGps = [4]map[string]float64{{"lon": 119.137911, "lat": 32.110117}, {"lon": 119.134857, "lat": 32.109321}, {"lon": 119.135216, "lat": 32.107395}, {"lon": 119.137875, "lat": 32.107395}}
var endWorkSpaceGps = [4]map[string]float64{{"lon": 119.140936, "lat": 32.104948}, {"lon": 119.143085, "lat": 32.105559}, {"lon": 119.131408, "lat": 32.105774}, {"lon": 119.125869, "lat": 32.105881}}
var excavatorNames = [4]string{"挖机01", "挖机02", "挖机03", "挖机04"}
var truckNames = [10]string{"自卸车01", "自卸车02", "自卸车03", "自卸车04", "自卸车05", "自卸车06", "自卸车07", "自卸车08", "自卸车09", "自卸车10"}
var sn = [10]string{"3402212212261", "3402212212262", "", "", "", "", "", "", "", ""}
var maxTruckNum = 2
var dispatchTimer = time.Tick(time.Second * 5)

func InitDispatch(cx *gin.Context) {
	go func() {
		for {
			select {
			case <-dispatchTimer:
				// 定时器到期，执行相应操作
				res := infra.RedisClient.Get(cx, DISPATCHKEY)
				if res.Err() == nil {
					var data model.Dispatch
					result, _ := res.Bytes()
					err := json.Unmarshal(result, &data)
					if err == nil {
						for i, platform := range data.Platforms {
							for j, truck := range platform.StartWorkSpace.Excavator.Trucks {
								if truck.Status == 1 {
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Status = 3
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Lon = platform.EndWorkSpace.Lon
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Lat = platform.EndWorkSpace.Lat
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Route = gpsSimulator(geo.NewPoint(platform.StartWorkSpace.Lat, platform.StartWorkSpace.Lon), geo.NewPoint(platform.EndWorkSpace.Lat, platform.EndWorkSpace.Lon))
									fmt.Println("车子正在去卸料")
								} else if truck.Status == 3 {
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Status = 1
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Lon = platform.StartWorkSpace.Lon
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Lat = platform.StartWorkSpace.Lat
									data.Platforms[i].StartWorkSpace.Excavator.Trucks[j].Route = gpsSimulator(geo.NewPoint(platform.EndWorkSpace.Lat, platform.EndWorkSpace.Lon), geo.NewPoint(platform.StartWorkSpace.Lat, platform.StartWorkSpace.Lon))
									fmt.Println("车子正在去装料")
								}
							}
						}
						redisData, err := json.Marshal(data)
						if err != nil {
							infra.Zaplog.Error("格式化数据失败：" + err.Error())
						}
						res := infra.RedisClient.Set(cx, DISPATCHKEY, redisData, 0)
						if res.Err() != nil {
							infra.Zaplog.Error("初始化失败：" + res.Err().Error())
						}
						infra.Zaplog.Info("车子正在装卸")
						fmt.Println("车子正在装卸")
					}
				}
			}
		}
	}()
	var platforms []model.Platform
	var excavators []model.Excavator
	var trucks []model.Truck
	var parking = model.Parking{
		Name:   "停车场",
		Lon:    119.142511,
		Lat:    32.110606,
		Radius: 100,
		Trucks: nil,
	}
	infra.RedisClient.Del(cx, DISPATCHKEY)
	min := 2
	max := 5
	platformNum = rand.Intn(max-min) + min
	// 初始化挖机和自卸车
	for i := 0; i < platformNum; i++ {
		excavators = append(excavators, model.Excavator{
			Id:              i + 1,
			Name:            excavatorNames[i],
			Status:          "normal",
			Trucks:          nil,
			CurrentTruckNum: 0,
			MaxTruckNum:     maxTruckNum,
		})
	}

	for i := 0; i < 10; i++ {
		trucks = append(trucks, model.Truck{
			Id:     i + 1,
			Name:   truckNames[i],
			Sn:     sn[i],
			Lon:    parking.Lon,
			Lat:    parking.Lat,
			Status: 0,
		})
	}

	// 初始化平台
	for i := 0; i < platformNum; i++ {
		platform := model.Platform{
			Id: i + 1,
			StartWorkSpace: model.StartWorkSpace{
				WorkSpace: model.WorkSpace{
					Id:     i + 1,
					Name:   startWorkSpaceNames[i],
					Lon:    startWorkSpaceGps[i]["lon"],
					Lat:    startWorkSpaceGps[i]["lat"],
					Radius: 100,
				},
			},
			EndWorkSpace: model.EndWorkSpace{
				WorkSpace: model.WorkSpace{
					Id:     i + 1,
					Name:   endWorkSpaceNames[i],
					Lon:    endWorkSpaceGps[i]["lon"],
					Lat:    endWorkSpaceGps[i]["lat"],
					Radius: 100,
				},
			},
		}
		excavator := excavators[i]
		excavator.Lon = platform.StartWorkSpace.Lon
		excavator.Lat = platform.StartWorkSpace.Lat
		trucksEntryNum := rand.Intn(maxTruckNum + 1)
		trucksEntry := trucks[:trucksEntryNum]
		trucks = trucks[trucksEntryNum:]

		for k := range trucksEntry {
			trucksEntry[k].Lat = platform.StartWorkSpace.Lat
			trucksEntry[k].Lon = platform.StartWorkSpace.Lon
			trucksEntry[k].Status = 1
			trucksEntry[k].Route = gpsSimulator(geo.NewPoint(parking.Lat, parking.Lon), geo.NewPoint(platform.StartWorkSpace.Lat, platform.StartWorkSpace.Lon))
		}
		excavator.CurrentTruckNum = trucksEntryNum
		excavator.Trucks = trucksEntry
		platform.StartWorkSpace.Excavator = excavator
		platforms = append(platforms, platform)
	}

	parking.Trucks = trucks
	data := model.Dispatch{
		Parking:   parking,
		Platforms: platforms,
	}
	redisData, err := json.Marshal(data)
	if err != nil {
		infra.Zaplog.Error("格式化数据失败：" + err.Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1001, "data": nil, "msg": err.Error()})
		return
	}
	res := infra.RedisClient.Set(cx, DISPATCHKEY, redisData, 0)
	if res.Err() != nil {
		infra.Zaplog.Error("初始化失败：" + res.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1002, "data": nil, "msg": res.Err().Error()})
		return
	}
	cx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": data})
	return
}

func GetDispatch(cx *gin.Context) {
	res := infra.RedisClient.Get(cx, DISPATCHKEY)
	if res.Err() != nil {
		infra.Zaplog.Error("获取平台数据失败：" + res.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1003, "data": nil, "msg": res.Err().Error()})
		return
	}
	var data model.Dispatch
	result, _ := res.Bytes()
	err := json.Unmarshal(result, &data)
	if err != nil {
		infra.Zaplog.Error("获取平台数据格式化失败：" + res.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1003, "data": nil, "msg": err.Error()})
		return
	}
	cx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": data, "msg": nil})
}

var mx sync.Mutex

func DispatchEdit(cx *gin.Context) {
	defer mx.Unlock()
	mx.Lock()
	infra.Zaplog.Info("Dispatch Edit Req Data:")
	var req request.DispatchEditRequest

	err := cx.ShouldBind(&req)
	if err != nil {
		infra.Zaplog.Error("调度参数错误：" + err.Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1003, "data": nil, "msg": err.Error()})
		return
	}
	res := infra.RedisClient.Get(cx, DISPATCHKEY)
	if res.Err() != nil {
		infra.Zaplog.Error("获取平台数据失败：" + res.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1003, "data": nil, "msg": res.Err().Error()})
		return
	}
	var data model.Dispatch
	result, _ := res.Bytes()
	err = json.Unmarshal(result, &data)
	if err != nil {
		infra.Zaplog.Error("获取平台数据格式化失败：" + res.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1003, "data": nil, "msg": err.Error()})
		return
	}

	var fromId int
	var truckReq model.Truck
	// 遍历，查看当前机械所在区域

	truckReq, _ = lo.Find(data.Parking.Trucks, func(truck model.Truck) bool {
		return truck.Id == req.Id
	})

	if truckReq.Id == req.Id {
		fromId = -1
	} else {
		//查询在工作区
	lookup:
		for _, platform := range data.Platforms {
			for _, truck := range platform.StartWorkSpace.Excavator.Trucks {
				if truck.Id == req.Id {
					truckReq = truck
					fromId = platform.Id
					break lookup
				}
			}
		}

		if truckReq.Id == 0 {
			infra.Zaplog.Error("当前平台机械不存在：" + res.Err().Error())
			cx.JSON(http.StatusOK, gin.H{"code": 1005, "data": nil, "msg": err.Error()})
			return
		}

	}

	switch {
	case req.ToPlatformId > 0 && fromId == -1:
		// 停车场->工作区
		data.Parking.Trucks = lo.FilterMap(data.Parking.Trucks, func(truck model.Truck, index int) (model.Truck, bool) {
			if truck.Id == req.Id {
				return model.Truck{}, false
			}
			return truck, true
		})
		for i, platform := range data.Platforms {
			if platform.Id == req.ToPlatformId {
				//判断平台机械是否已经满了
				if platform.StartWorkSpace.Excavator.CurrentTruckNum >= platform.StartWorkSpace.Excavator.MaxTruckNum && req.Type != 1 {
					infra.Zaplog.Error("平台已满")
					cx.JSON(http.StatusOK, gin.H{"code": 1008, "data": nil, "msg": "平台已满，调度失败"})
					return
				}
				truckReq.Route = gpsSimulator(geo.NewPoint(truckReq.Lat, truckReq.Lon), geo.NewPoint(platform.StartWorkSpace.Lat, platform.StartWorkSpace.Lon))
				truckReq.Lon = platform.StartWorkSpace.Lon
				truckReq.Lat = platform.StartWorkSpace.Lat
				truckReq.Status = 1
				data.Platforms[i].StartWorkSpace.Excavator.Trucks = append(platform.StartWorkSpace.Excavator.Trucks, truckReq)
				data.Platforms[i].StartWorkSpace.Excavator.CurrentTruckNum += 1
				break
			}
		}
	case req.ToPlatformId < 0 && fromId > 0:
		// 工作区->停车场
		for i, platform := range data.Platforms {
			if platform.Id == fromId {
				data.Platforms[i].StartWorkSpace.Excavator.Trucks = lo.FilterMap(platform.StartWorkSpace.Excavator.Trucks, func(truck model.Truck, index int) (model.Truck, bool) {
					if truck.Id == req.Id {
						return model.Truck{}, false
					}
					return truck, true
				})
				data.Platforms[i].StartWorkSpace.Excavator.CurrentTruckNum -= 1
				break
			}
		}
		truckReq.Route = gpsSimulator(geo.NewPoint(truckReq.Lat, truckReq.Lon), geo.NewPoint(data.Parking.Lat, data.Parking.Lon))
		truckReq.Lon = data.Parking.Lon
		truckReq.Lat = data.Parking.Lat
		truckReq.Status = 0
		data.Parking.Trucks = append(data.Parking.Trucks, truckReq)
	case req.ToPlatformId > 0 && fromId > 0 && req.ToPlatformId != fromId:
		// 工作区->工作区
		for i, platform := range data.Platforms {
			if platform.Id == fromId {
				data.Platforms[i].StartWorkSpace.Excavator.Trucks = lo.FilterMap(platform.StartWorkSpace.Excavator.Trucks, func(truck model.Truck, index int) (model.Truck, bool) {
					if truck.Id == req.Id {
						return model.Truck{}, false
					}
					return truck, true
				})
				data.Platforms[i].StartWorkSpace.Excavator.CurrentTruckNum -= 1
				break
			}
		}
		for i, platform := range data.Platforms {
			if platform.Id == req.ToPlatformId {
				if platform.StartWorkSpace.Excavator.CurrentTruckNum >= platform.StartWorkSpace.Excavator.MaxTruckNum && req.Type != 1 {
					infra.Zaplog.Error("平台已满")
					cx.JSON(http.StatusOK, gin.H{"code": 1009, "data": nil, "msg": "平台已满，调度失败"})
					return
				}
				truckReq.Route = gpsSimulator(geo.NewPoint(truckReq.Lat, truckReq.Lon), geo.NewPoint(platform.StartWorkSpace.Lat, platform.StartWorkSpace.Lon))
				truckReq.Lon = platform.StartWorkSpace.Lon
				truckReq.Lat = platform.StartWorkSpace.Lat
				truckReq.Status = 1
				data.Platforms[i].StartWorkSpace.Excavator.CurrentTruckNum += 1
				data.Platforms[i].StartWorkSpace.Excavator.Trucks = append(platform.StartWorkSpace.Excavator.Trucks, truckReq)
				break
			}
		}
	default:
		infra.Zaplog.Info("场景不处理")
		cx.JSON(http.StatusOK, gin.H{"code": 1005, "data": nil, "msg": "场景不处理"})
		return
	}

	redisData, err := json.Marshal(data)
	if err != nil {
		infra.Zaplog.Error("调度格式化数据失败：" + err.Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1001, "data": nil, "msg": err.Error()})
		return
	}

	r := infra.RedisClient.Set(cx, DISPATCHKEY, redisData, 0)
	if r.Err() != nil {
		infra.Zaplog.Error("调度保存失败：" + r.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1002, "data": nil, "msg": r.Err().Error()})
		return
	}

	cx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": data, "msg": nil})
}

func gpsSimulator(start *geo.Point, end *geo.Point) (res []model.Route) {
	// 两个GPS点之间的距离
	bearing := start.BearingTo(end) // 两个GPS点之间的方向角
	// 每500米生成一个GPS点
	distance := start.GreatCircleDistance(end) / 10 // 两个GPS点之间的距离
	res = append(res, model.Route{Lat: start.Lat(), Lon: start.Lng()})
	for i := 1; i < 10; i++ {
		// 计算每个中间点的经纬度
		mp := pointOnBearing(start, bearing, float64(i)*distance)
		res = append(res, model.Route{Lat: mp.Lat(), Lon: mp.Lng()})
	}

	return res
}

func pointOnBearing(p *geo.Point, bearing, distance float64) *geo.Point {
	// 计算新GPS点的经度和纬度
	lat1 := p.Lat() * math.Pi / 180.0
	lon1 := p.Lng() * math.Pi / 180.0
	b := bearing * math.Pi / 180.0
	d := distance / 6378.1 // 地球半径约为6378.1千米
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(d) + math.Cos(lat1)*math.Sin(d)*math.Cos(b))
	lon2 := lon1 + math.Atan2(math.Sin(b)*math.Sin(d)*math.Cos(lat1), math.Cos(d)-math.Sin(lat1)*math.Sin(lat2))
	// 将经度和纬度转换回弧度制
	lat2 = lat2 * 180.0 / math.Pi
	lon2 = lon2 * 180.0 / math.Pi
	return geo.NewPoint(lat2, lon2)
}

func Notify(cx *gin.Context) {
	var notifies request.NotifyRequest
	client := resty.New()
	err := cx.ShouldBind(&notifies)
	if err != nil {
		infra.Zaplog.Error("调度参数绑定失败：" + err.Error())
	}
	var wg sync.WaitGroup
	for _, machine := range notifies.Machines {
		wg.Add(1)
		go func(machine request.DispatchMachine) {
			defer wg.Done()
			res, err := client.R().
				EnableTrace().
				SetQueryParams(map[string]string{
					"id":      machine.Sn,
					"token":   config.AppConfig.Sound.Token,
					"version": strconv.Itoa(config.AppConfig.Sound.Version),
					"message": machine.Message,
				}).
				Get(config.AppConfig.Sound.Url)
			if err != nil {
				infra.Zaplog.Error("调度失败：" + err.Error())
			}

			infra.Zaplog.Info("调度结果：" + string(res.Body()))
			var result response.SoundNotifyResponse
			err = json.Unmarshal(res.Body(), &result)
			if err != nil {
				infra.Zaplog.Error("数据解析失败：" + err.Error())
			}
		}(machine)
	}

	wg.Wait()
	cx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "ok"})
}

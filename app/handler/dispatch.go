package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
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
var endWorkSpaceGps = [4]map[string]float64{{"lon": 119.140936, "lat": 32.104948}, {"lon": 119.143085, "lat": 32.105559}, {"lon": 119.131408, "lat": 32.105774}, {"lon": 119.130025, "lat": 32.105881}}
var excavatorNames = [4]string{"挖机01", "挖机02", "挖机03", "挖机04"}
var truckNames = [4]string{"自卸车01", "自卸车02", "自卸车03", "自卸车04"}
var sn = [4]string{"3402212212261", "3402212212262", "", ""}
var maxTruckNum = 4

func InitDispatch(cx *gin.Context) {
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

		trucks = append(trucks, model.Truck{
			Id:   i + 1,
			Name: truckNames[i],
			Sn:   sn[i],
			Lon:  parking.Lon,
			Lat:  parking.Lat,
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
					Radius: 10,
				},
			},
			EndWorkSpace: model.EndWorkSpace{
				WorkSpace: model.WorkSpace{
					Id:     i + 1,
					Name:   endWorkSpaceNames[i],
					Lon:    endWorkSpaceGps[i]["lon"],
					Lat:    endWorkSpaceGps[i]["lat"],
					Radius: 10,
				},
			},
		}
		excavator := excavators[i]
		excavator.Lon = platform.StartWorkSpace.Lon
		excavator.Lat = platform.StartWorkSpace.Lat
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
	res := infra.RedisClient.SetNX(cx, DISPATCHKEY, redisData, 0)
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

func DispatchEdit(cx *gin.Context) {
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
				break
			}
		}
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

	r := infra.RedisClient.SetNX(cx, DISPATCHKEY, redisData, 0)
	if r.Err() != nil {
		infra.Zaplog.Error("调度保存失败：" + r.Err().Error())
		cx.JSON(http.StatusOK, gin.H{"code": 1002, "data": nil, "msg": r.Err().Error()})
		return
	}

	cx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": data, "msg": nil})
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

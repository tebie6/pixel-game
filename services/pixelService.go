package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/tools/lock"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type PixelService struct{}

// r表示receive（接受者） 更新像素
func (r *PixelService) SavePixel(x int64, y int64, color int64, uid int64) (requiredLogin bool, err error) {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()
	//reply, _ := conn.Do("HGET", "pixel_operation_count", fmt.Sprint("uid:", uid))
	//operationCount, _ := redis.Int64(reply, err)

	db := models.GetDbInst()

	// 创建记录
	gamePixelRecord := models.GamePixelRecord{
		X:         x,
		Y:         y,
		Color:     color,
		Uid:       uid,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	err = db.Create(&gamePixelRecord).Error

	// 存储到缓存
	conn.Do("HSET", "pixel_hash", fmt.Sprint(x, ":", y), color)

	// 释放生成画布图片的锁，让画布图片更新
	lock.ReleaseLock(conn, "generate_canvas_image_lock")

	return false, nil
}

// GetList 获取画布列表
func (r *PixelService) GetList() (res map[string]int64, err error) {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	//reply, err := conn.Do("HGETALL", "pixel_hash1")
	//if err != nil {
	//	return nil, errors.New("服务器内部错误")
	//}
	//res, _ = redis.StringMap(reply, err)

	// 初始化游标和用于存储结果的map
	var cursor uint64
	res = make(map[string]int64)

	// 使用for循环进行多批次请求
	for {
		// 执行HSCAN命令，从pixel_hash哈希表中获取一批数据
		values, err := redis.Values(conn.Do("HSCAN", "pixel_hash", cursor, "COUNT", 1000))
		if err != nil {
			// 如果HSCAN执行出错，返回错误
			return nil, errors.New("服务器内部错误")
		}

		// 解析HSCAN返回的数据，包括新的游标和键值对
		var keysValues []string
		_, err = redis.Scan(values, &cursor, &keysValues)
		if err != nil {
			// 如果解析出错，返回错误
			return nil, errors.New("服务器内部错误")
		}

		// 遍历获取的键值对，并将它们添加到结果map中
		for i := 0; i < len(keysValues); i += 2 {
			res[keysValues[i]], _ = strconv.ParseInt(keysValues[i+1], 10, 64)
		}
		keysValues = nil

		// 检查游标，如果为0，则表示HSCAN遍历完成
		if cursor == 0 {
			break
		}
	}

	return
}

// RepairContent 修复画布内容
func (r *PixelService) RepairContent() (err error) {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	db := models.GetDbInst()
	var gamePixelContent []models.GamePixelRecord
	pageSize := 10000 // 每批处理的记录数
	conn.Do("DEL", "pixel_hash")
	var lastID int64 = 0 // 初始化为最小的 id

	for {
		fmt.Println("lastID", lastID)
		err = db.Model(&models.GamePixelRecord{}).
			Where("id > ? AND status = 1", lastID).
			Order("id asc").
			Limit(pageSize).
			Find(&gamePixelContent).Error

		// 相同的错误处理和业务逻辑
		for _, item := range gamePixelContent {
			key := fmt.Sprintf("%d:%d", item.X, item.Y)
			// 存储到缓存
			conn.Do("HSET", "pixel_hash", key, item.Color)
		}

		// 更新 lastID 为当前批次的最后一条记录的 id
		if len(gamePixelContent) > 0 {
			lastID = gamePixelContent[len(gamePixelContent)-1].Id
		} else {
			break
		}
	}

	return
}

// Message 定义了聊天消息的结构
type Message struct {
	Username string `json:"username"`
	Msg      string `json:"msg"`
}

// SaveChat 保存聊天消息
func (r *PixelService) SaveChat(username string, msg string) error {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	// 创建一个新的消息实例
	message := Message{
		Username: username,
		Msg:      msg,
	}

	// 将消息结构体转换为JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 将JSON消息推送到Redis列表
	listKey := "chatList"
	if _, err := conn.Do("RPUSH", listKey, messageJSON); err != nil {
		return err
	}

	return nil
}

// GetChatList 获取聊天列表
func (r *PixelService) GetChatList() (res []Message, err error) {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	// 列表的键名
	key := "chatList"

	// 使用LRANGE命令获取最近30条记录
	messages, err := redis.Strings(conn.Do("LRANGE", key, -30, -1))
	if err != nil {
		return res, nil
	}

	// 解析每条消息
	for _, msg := range messages {
		var message Message
		err = json.Unmarshal([]byte(msg), &message)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			continue
		}
		res = append(res, message)
	}

	// 追加系统公告
	res = append(res, Message{
		Username: "系统提醒",
		Msg:      "自动化脚本请移步空白处！！！谢谢！",
	})
	res = append(res, Message{
		Username: "系统提醒",
		Msg:      "请避免使用敏感或不适当的语言，让我们一起维护友好的交流环境。谢谢合作！😊",
	})

	return
}

// GenerateCanvasImage 生成画布图片
func (r *PixelService) GenerateCanvasImage() (err error) {

	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	lockKey := "generate_canvas_image_lock"
	// 获得锁
	if !lock.ObtainLock(conn, lockKey, 3600*time.Second) {
		return
	}

	// 创建一个宽度为 1000，高度为 1000 的图像
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))

	colorList := `{"1":"0,0,0","2":"85,85,85","3":"136,136,136","4":"205,205,205","0":"255,255,255","5":"255,213,188","6":"255,183,131","7":"182,109,61","8":"119,67,31","9":"252,117,16","10":"252,168,14","11":"253,232,23","12":"255,244,145","13":"190,255,64","14":"112,221,19","15":"49,161,23","16":"50,182,159","17":"136,255,243","18":"36,181,254","19":"18,92,199","20":"38,41,96","21":"139,47,168","22":"255,89,239","23":"255,169,217","24":"255,100,116","25":"240,37,35","26":"177,18,6","27":"116,12,0","100":"0,0,0","101":"105,105,105","102":"128,128,128","103":"169,169,169","104":"192,192,192","105":"211,211,211","106":"220,220,220","107":"245,245,245","108":"255,255,255","109":"128,0,0","110":"139,0,0","111":"178,34,34","112":"165,42,42","113":"255,0,0","114":"205,92,92","115":"188,143,143","116":"240,128,128","117":"255,250,250","118":"250,128,114","119":"255,228,225","120":"255,99,71","121":"233,150,122","122":"255,69,0","123":"255,127,80","124":"255,160,122","125":"160,82,45","126":"255,245,238","127":"139,69,19","128":"210,105,30","129":"244,164,96","130":"255,218,185","131":"205,133,63","132":"250,240,230","133":"255,140,0","134":"255,228,196","135":"222,184,135","136":"210,180,140","137":"250,235,215","138":"255,222,173","139":"255,235,205","140":"255,239,213","141":"255,165,0","142":"255,228,181","143":"245,222,179","144":"253,245,230","145":"255,250,240","146":"218,165,32","147":"255,248,220","148":"255,215,0","149":"240,230,140","150":"238,232,170","151":"255,250,205","152":"189,183,107","153":"128,128,0","154":"255,255,0","155":"255,255,224","156":"255,255,240","157":"250,250,210","158":"245,245,220","159":"85,107,47","160":"173,255,47","161":"124,252,0","162":"127,255,0","163":"0,100,0","164":"0,128,0","165":"34,139,34","166":"0,255,0","167":"50,205,50","168":"143,188,143","169":"152,251,152","170":"144,238,144","171":"240,255,240","172":"46,139,87","173":"60,179,113","174":"245,255,250","175":"0,255,127","176":"0,250,154","177":"127,255,170","178":"64,224,208","179":"32,178,170","180":"72,209,204","181":"0,128,128","182":"0,139,139","183":"47,79,79","184":"0,206,209","185":"212,242,231","186":"0,255,255","187":"175,238,238","188":"225,255,255","189":"240,255,255","190":"95,158,160","191":"176,224,230","192":"173,216,230","193":"0,191,255","194":"135,206,235","195":"135,206,250","196":"70,130,180","197":"240,248,255","198":"30,144,255","199":"112,128,144","200":"119,136,153","201":"176,196,222","202":"100,149,237","203":"65,105,225","204":"0,0,128","205":"0,0,139","206":"25,25,112","207":"0,0,205","208":"0,0,255","209":"248,248,255","210":"230,230,250","211":"72,61,139","212":"106,90,205","213":"123,104,238","214":"147,112,219","215":"138,43,226","216":"75,0,130","217":"153,50,204","218":"148,0,211","219":"186,85,211","220":"128,0,128","221":"139,0,139","222":"255,0,255","223":"255,0,255","224":"238,130,238","225":"221,160,221","226":"216,191,216","227":"218,112,214","228":"199,21,133","229":"255,20,147","230":"255,105,180","231":"219,112,147","232":"255,240,245","233":"220,20,60","234":"255,192,203","235":"255,182,193"}`
	var colors map[string]string
	err = json.Unmarshal([]byte(colorList), &colors)
	if err != nil {
		return errors.New("无法解析颜色列表")
	}

	// 初始化游标和用于存储结果的map
	var cursor uint64

	// 使用for循环进行多批次请求
	for {
		// 执行HSCAN命令，从pixel_hash哈希表中获取一批数据
		values, err := redis.Values(conn.Do("HSCAN", "pixel_hash", cursor, "COUNT", 1000))
		if err != nil {
			// 如果HSCAN执行出错，返回错误
			return errors.New("服务器内部错误")
		}

		// 解析HSCAN返回的数据，包括新的游标和键值对
		var keysValues []string
		_, err = redis.Scan(values, &cursor, &keysValues)
		if err != nil {
			// 如果解析出错，返回错误
			return errors.New("服务器内部错误")
		}

		// 遍历获取的键值对，并将它们添加到结果map中
		for i := 0; i < len(keysValues); i += 2 {
			coords := strings.Split(keysValues[i], ":")
			if len(coords) != 2 {
				continue // 错误的坐标格式
			}

			x, errX := strconv.Atoi(coords[0])
			y, errY := strconv.Atoi(coords[1])
			if errX != nil || errY != nil {
				continue // 无效的坐标
			}

			colorID := keysValues[i+1]
			colorValue, exists := colors[colorID]
			if !exists {
				continue // 未知的颜色 ID
			}

			R, G, B := parseColor(colorValue)
			img.Set(x, y, color.RGBA{R: R, G: G, B: B, A: 255})
		}
		keysValues = nil

		// 检查游标，如果为0，则表示HSCAN遍历完成
		if cursor == 0 {
			break
		}
	}

	// 临时文件的作用是防止图片未处理完被加载使用
	tempFilePath := "./frontend-lib/static/pixel/img/temp_canvas.png" // 临时文件路径
	finalFilePath := "./frontend-lib/static/pixel/img/canvas.png"     // 最终文件路径

	// 保存图像
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return errors.New("无法创建图像文件")
	}
	defer tempFile.Close()

	err = png.Encode(tempFile, img)
	if err != nil {
		return errors.New("无法保存图像")
	}

	// 将临时文件重命名为最终文件
	err = os.Rename(tempFilePath, finalFilePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("图像更新成功")

	return
}

// parseColor 解析颜色字符串为 RGB
func parseColor(s string) (r, g, b uint8) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return 0, 0, 0 // 或者返回一个错误
	}

	ri, _ := strconv.ParseUint(parts[0], 10, 8)
	gi, _ := strconv.ParseUint(parts[1], 10, 8)
	bi, _ := strconv.ParseUint(parts[2], 10, 8)

	return uint8(ri), uint8(gi), uint8(bi)
}

// ErrorReporting 错误上报
func (r *PixelService) ErrorReporting(message string, source string, lineno string, colno string, stack string, accessToken string) (err error) {

	db := models.GetDbInst()

	// 验证token
	userInfo, err := models.GetUserByAccessToken(accessToken)
	if err != nil {
		return errors.New("非法请求 10001")
	}

	if userInfo == nil {
		return errors.New("非法请求 10002")

	}

	// 禁用用户
	if userInfo.Status == 0 {
		return errors.New("非法请求 10003")
	}

	// 组装数据
	gameErrorContent := models.GameErrorContent{}
	gameErrorContent.Message = message
	gameErrorContent.Source = source
	gameErrorContent.Lineno = lineno
	gameErrorContent.Colno = colno
	gameErrorContent.Stack = stack
	gameErrorContent.Uid = userInfo.Id
	gameErrorContent.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	err = db.Create(&gameErrorContent).Error
	if err != nil {
		return err
	}

	return nil
}

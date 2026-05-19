package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[PHOTO-SERVICE] ", log.LstdFlags|log.Lshortfile)

	// Load config (dev/prod based on ENV variable)
	if err := loadConfig(); err != nil {
		logger.Fatalf("配置加载失败: %v", err)
	}
	logger.Printf("环境: %s, 端口: %s", cfg.Env, cfg.Port)

	if err := initDB(); err != nil {
		logger.Fatalf("数据库初始化失败: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func jsonResp(w http.ResponseWriter, success bool, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": msg,
		"data":    data,
	})
}

// ==================== generatePhoto — 核心全流程 ====================

var photoSpecs = map[string]map[string]interface{}{
	"cn_passport":       {"name": "中国护照", "w": 33, "h": 48, "dpi": 300, "bg": "#FFFFFF"},
	"cn_id":             {"name": "中国身份证", "w": 26, "h": 32, "dpi": 300, "bg": "#FFFFFF"},
	"cn_one_inch":       {"name": "一寸照片", "w": 25, "h": 35, "dpi": 300, "bg": "#FFFFFF"},
	"cn_two_inch":       {"name": "二寸照片", "w": 35, "h": 49, "dpi": 300, "bg": "#FFFFFF"},
	"cn_small_one_inch": {"name": "小一寸", "w": 22, "h": 32, "dpi": 300, "bg": "#FFFFFF"},
	"cn_small_two_inch": {"name": "小二寸", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"cn_big_two_inch":   {"name": "大二寸", "w": 35, "h": 53, "dpi": 300, "bg": "#FFFFFF"},
	"cn_one_inch_blue":  {"name": "一寸蓝底", "w": 25, "h": 35, "dpi": 300, "bg": "#1E90FF"},
	"cn_one_inch_red":   {"name": "一寸红底", "w": 25, "h": 35, "dpi": 300, "bg": "#FF0000"},
	"cn_two_inch_blue":  {"name": "二寸蓝底", "w": 35, "h": 49, "dpi": 300, "bg": "#1E90FF"},
	"cn_two_inch_red":   {"name": "二寸红底", "w": 35, "h": 49, "dpi": 300, "bg": "#FF0000"},
	"cn_driver_license": {"name": "驾驶证", "w": 22, "h": 32, "dpi": 300, "bg": "#FFFFFF"},
	"us_passport":       {"name": "美国护照", "w": 51, "h": 51, "dpi": 300, "bg": "#FFFFFF"},
	"us_visa":           {"name": "美国签证", "w": 51, "h": 51, "dpi": 300, "bg": "#FFFFFF"},
	"jp_passport":       {"name": "日本护照", "w": 45, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"jp_visa":           {"name": "日本签证", "w": 45, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"jp_visa_rect":      {"name": "日本签证（矩形）", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"schengen_visa":     {"name": "申根签证", "w": 35, "h": 45, "dpi": 300, "bg": "#F0F0F0"},
	"uk_passport":       {"name": "英国护照", "w": 35, "h": 45, "dpi": 300, "bg": "#F0F0F0"},
	"de_passport":       {"name": "德国护照", "w": 35, "h": 45, "dpi": 300, "bg": "#E8E8E8"},
	"fr_passport":       {"name": "法国护照", "w": 35, "h": 45, "dpi": 300, "bg": "#E8E8E8"},
	"ca_passport":       {"name": "加拿大护照", "w": 50, "h": 70, "dpi": 300, "bg": "#FFFFFF"},
	"au_passport":       {"name": "澳大利亚护照", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"kr_passport":       {"name": "韩国护照", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"sg_passport":       {"name": "新加坡护照", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"standard_35x45":    {"name": "标准签证照", "w": 35, "h": 45, "dpi": 300, "bg": "#FFFFFF"},
	"standard_33x48":    {"name": "标准护照照", "w": 33, "h": 48, "dpi": 300, "bg": "#FFFFFF"},
}

func generatePhoto(w http.ResponseWriter, r *http.Request) {
	logger.Printf("[%s] generate-photo from %s", r.Method, r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	spec := r.FormValue("spec")
	gender := r.FormValue("gender")
	if spec == "" {
		spec = "standard_35x45"
	}
	if gender == "" {
		gender = "female"
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		jsonResp(w, false, "文件上传失败", nil)
		return
	}
	defer file.Close()

	cfg2, ok := photoSpecs[spec]
	if !ok {
		cfg2 = photoSpecs["standard_35x45"]
	}
	logger.Printf("开始生成证件照: %s, 性别: %s", cfg2["name"], gender)

	// ── Step 1: Upload original image to COS ──
	cosKey := generateCOSKey("uploads", header.Filename)
	data, _ := io.ReadAll(file)
	origCOSURL, err := cosUpload(data, cosKey, "image/jpeg")
	if err != nil {
		logger.Printf("COS上传失败: %v", err)
		jsonResp(w, false, "COS上传失败", nil)
		return
	}
	logger.Printf("原始图COS: %s", origCOSURL)

	// ── Step 2: FaceChain generates professional portrait ──
	templateURL := cfg.FaceChainTemplateFemale
	if gender == "male" {
		templateURL = cfg.FaceChainTemplateMale
	}
	facechainResultURL, fcErr := callFaceChain(origCOSURL, gender, templateURL)
	if fcErr != nil {
		logger.Printf("FaceChain失败, 降级使用原图: %v", fcErr)
		facechainResultURL = origCOSURL
	}

	// ── Step 3: Download FaceChain result ──
	resultData, dlErr := downloadImage(facechainResultURL)
	if dlErr != nil {
		logger.Printf("下载FaceChain结果失败: %v", dlErr)
		resultData = data // fallback to original
	}

	// Save to temp file for Hivision processing
	tmpFile, _ := os.CreateTemp("", "fc-result-*.jpg")
	if tmpFile != nil {
		tmpFile.Write(resultData)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())
	}

	// ── Step 4: HivisionIDPhotos processing ──
	var finalData []byte
	if tmpFile != nil {
		hivisionPath, hvErr := callHivisionIDPhoto(tmpFile.Name(), cfg2)
		if hvErr != nil {
			logger.Printf("Hivision失败，返回FaceChain原图: %v", hvErr)
			finalData = resultData
		} else {
			finalData, _ = os.ReadFile(hivisionPath)
			os.Remove(hivisionPath)
		}
	} else {
		finalData = resultData
	}

	// ── Step 5: Upload final result to COS ──
	resultCOSKey := generateCOSKey("results", "final_result.jpg")
	finalURL, err := cosUpload(finalData, resultCOSKey, "image/jpeg")
	if err != nil {
		logger.Printf("COS上传最终结果失败: %v", err)
		// Fallback: return base64
		b64 := base64.StdEncoding.EncodeToString(finalData)
		jsonResp(w, true, "ok (no COS)", map[string]interface{}{
			"image": "data:image/jpeg;base64," + b64,
			"spec":  cfg2["name"],
		})
		return
	}
	logger.Printf("最终结果COS: %s", finalURL)

	// ── Step 6: Log generation ──
	if _, err := db.Exec(`INSERT INTO generation_logs (spec_name, gender, source_file, result_file, status) VALUES (?, ?, ?, ?, 1)`,
		cfg2["name"], gender, origCOSURL, finalURL); err != nil {
		logger.Printf("记录generation_logs失败: %v", err)
	}

	jsonResp(w, true, "ok", map[string]interface{}{
		"image": finalURL,
		"spec":  cfg2["name"],
	})
}

// ==================== HivisionIDPhotos ====================

func callHivisionIDPhoto(imagePath string, specCfg map[string]interface{}) (string, error) {
	width := specCfg["w"].(int)
	height := specCfg["h"].(int)
	dpi := 300
	if dpiVal, ok := specCfg["dpi"].(int); ok {
		dpi = dpiVal
	}
	widthPx := int(float64(width) / 25.4 * float64(dpi))
	heightPx := int(float64(height) / 25.4 * float64(dpi))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("input_image", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("创建表单失败: %v", err)
	}
	io.Copy(part, file)

	writer.WriteField("height", fmt.Sprintf("%d", heightPx))
	writer.WriteField("width", fmt.Sprintf("%d", widthPx))
	writer.WriteField("dpi", fmt.Sprintf("%d", dpi))
	writer.WriteField("hd", "true")
	writer.WriteField("whitening_strength", "0")
	writer.WriteField("head_measure_ratio", "0.2")
	writer.WriteField("head_height_ratio", "0.45")
	writer.WriteField("top_distance_max", "0.12")
	writer.WriteField("top_distance_min", "0.10")
	writer.WriteField("face_detect_model", "mtcnn")
	writer.WriteField("human_matting_model", "modnet_photographic_portrait_matting")
	writer.Close()

	resp, err := http.Post(cfg.HivisionURL, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("Hivision调用失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Hivision HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析Hivision响应失败: %v", err)
	}

	if status, ok := result["status"].(bool); !ok || !status {
		msg := "人脸检测失败，请上传单人正面照"
		if errMsg, ok := result["message"].(string); ok {
			msg = errMsg
		}
		return "", fmt.Errorf(msg)
	}

	// Face orientation validation (passport/visa requires frontal face)
	if faceData, ok := result["face"].(map[string]interface{}); ok {
		if yaw, ok := faceData["yaw_angle"].(float64); ok && (yaw > 15 || yaw < -15) {
			return "", fmt.Errorf("检测到侧脸，偏航角%.0f°，请上传正面照片", yaw)
		}
		if pitch, ok := faceData["pitch_angle"].(float64); ok && (pitch > 20 || pitch < -20) {
			return "", fmt.Errorf("检测到低头或仰头，俯仰角%.0f°，请上传正面平视照片", pitch)
		}
		logger.Printf("人脸角度: yaw=%.1f, pitch=%.1f, roll=%.1f",
			faceData["yaw_angle"], faceData["pitch_angle"], faceData["roll_angle"])
	}

	resultImageBase64, ok := result["image_base64_standard"].(string)
	if !ok {
		return "", fmt.Errorf("Hivision返回格式错误")
	}
	if strings.HasPrefix(resultImageBase64, "data:image") {
		parts := strings.SplitN(resultImageBase64, ",", 2)
		if len(parts) == 2 {
			resultImageBase64 = parts[1]
		}
	}

	// Apply background color from spec (before JPEG encoding loses alpha)
	if bgColor, ok := specCfg["bg"].(string); ok && bgColor != "" {
		bgColor = strings.TrimPrefix(bgColor, "#")
		if bgColor != "FFFFFF" && bgColor != "ffffff" {
			coloredB64, bgErr := callHivisionAddBackground(resultImageBase64, bgColor, dpi)
			if bgErr != nil {
				logger.Printf("添加背景色失败, 使用白色: %v", bgErr)
			} else {
				resultImageBase64 = coloredB64
				// Strip prefix again since add_background returns with it
				if strings.HasPrefix(resultImageBase64, "data:image") {
					parts := strings.SplitN(resultImageBase64, ",", 2)
					if len(parts) == 2 {
						resultImageBase64 = parts[1]
					}
				}
			}
		}
	}

	imageBytes, err := base64.StdEncoding.DecodeString(resultImageBase64)
	if err != nil {
		return "", fmt.Errorf("解码Hivision图片失败: %v", err)
	}

	outputPath := strings.Replace(imagePath, filepath.Ext(imagePath), "_processed.jpg", 1)
	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("保存Hivision结果失败: %v", err)
	}

	logger.Printf("Hivision完成: %dx%d px", widthPx, heightPx)
	return outputPath, nil
}

// callHivisionAddBackground fills the transparent background with a solid color.
func callHivisionAddBackground(imageBase64 string, color string, dpi int) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("input_image_base64", imageBase64)
	writer.WriteField("color", color)
	writer.WriteField("dpi", fmt.Sprintf("%d", dpi))
	writer.WriteField("render", "0") // pure_color
	writer.Close()

	addBgURL := strings.Replace(cfg.HivisionURL, "/idphoto", "/add_background", 1)
	resp, err := http.Post(addBgURL, writer.FormDataContentType(), body)
	if err != nil {
		return "", fmt.Errorf("add_background调用失败: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析add_background响应失败: %v", err)
	}

	if status, ok := result["status"].(bool); !ok || !status {
		return "", fmt.Errorf("add_background失败")
	}

	bgBase64, ok := result["image_base64"].(string)
	if !ok {
		return "", fmt.Errorf("add_background返回格式错误")
	}
	return bgBase64, nil
}

// listVideos returns a JSON list of video files from COS video-feed/ directory.
func listVideos(w http.ResponseWriter, r *http.Request) {
	keys, err := cosListObjects("video-feed/", 100)
	if err != nil {
		jsonResp(w, false, "列表获取失败", nil)
		return
	}
	// Strip prefix from keys
	var files []string
	for _, k := range keys {
		files = append(files, k[len("video-feed/"):])
	}
	if files == nil {
		files = []string{}
	}
	cosBase := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/video-feed/", cfg.COSBucket, cfg.COSRegion)
	jsonResp(w, true, "ok", map[string]interface{}{
		"videos":  files,
		"cosBase": cosBase,
	})
}

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// ==================== Main ====================

func main() {
	logger.Printf("========================================")
	logger.Printf("Photo ID Service v2.0 - All Go")
	logger.Printf("环境: %s  端口: %s", cfg.Env, cfg.Port)
	logger.Printf("存储: COS %s/%s", cfg.COSBucket, cfg.COSRegion)
	logger.Printf("AI引擎: FaceChain + HivisionIDPhotos")
	logger.Printf("========================================")

	http.HandleFunc("/api/generate-photo", corsMiddleware(generatePhoto))
	http.HandleFunc("/api/video-list", corsMiddleware(listVideos))

	logger.Printf("服务就绪")
	logger.Fatal(http.ListenAndServe(cfg.Port, nil))
}

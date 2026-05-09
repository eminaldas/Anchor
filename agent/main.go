package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"golang.org/x/sys/windows/registry"
)

// ──────────────────── Config ────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	serverURL     = getEnv("SENTINEL_SERVER_URL", "http://10.0.11.10:8081")
	reportPath    = "/api/v1/report"
	resultPath    = "/api/v1/commands/%s/result"
	heartbeatSec  = 10
	deepAuditSec  = 300 // 5 dakika
)

// ──────────────────── Models ────────────────────

type Finding struct {
	AssetID     string `json:"asset_id"`
	CheckType   string `json:"check_type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type HeartbeatPayload struct {
	CPU             string `json:"cpu"`
	RAM             string `json:"ram"`
	Uptime          uint64 `json:"uptime"`
	DefenderEnabled bool   `json:"defender_enabled"`
}

type ServerCommand struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Params string `json:"params"`
}

type ReportResponse struct {
	Message  string          `json:"message"`
	Commands []ServerCommand `json:"commands"`
}

type CommandResult struct {
	Status string `json:"status"` // completed, failed
	Result string `json:"result"`
}

// ──────────────────── Deep Audit Models ────────────────────

type FileHashInfo struct {
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	SizeBytes    int64  `json:"size_bytes"`
	LastModified string `json:"last_modified"`
}

type SecurityEventInfo struct {
	EventID         int    `json:"event_id"`
	TimeCreated     string `json:"time_created"`
	TargetUserName  string `json:"target_user_name"`
	LogonType       int    `json:"logon_type"`
	IpAddress       string `json:"ip_address"`
	WorkstationName string `json:"workstation_name"`
}

// ──────────────────── Main ────────────────────

func main() {
	hInfo, _ := host.Info()
	uniqueID := fmt.Sprintf("%s_%s", hInfo.Hostname, hInfo.HostID)

	fmt.Printf("Sentinel Agent Başlatıldı | Kimlik: %s\n", uniqueID)
	fmt.Printf("Sunucu: %s\n", serverURL)

	// 1. Yazılım envanteri (tek seferlik)
	fmt.Println("Yazılım envanteri toplanıyor...")
	softData, _ := json.Marshal(getInstalledSoftware())
	sendReport(Finding{
		AssetID:     uniqueID,
		CheckType:   "Software_Inventory",
		Severity:    "Info",
		Description: string(softData),
	}, uniqueID)

	// 2. Deep Audit döngüsü (5 dakikada bir)
	go func() {
		// İlk çalıştırmada hemen yap
		runDeepAudit(uniqueID)

		ticker := time.NewTicker(time.Duration(deepAuditSec) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			runDeepAudit(uniqueID)
		}
	}()

	// 3. Heartbeat döngüsü
	fmt.Println("Heartbeat başlatıldı...")
	for {
		payload := buildHeartbeat()
		payloadJSON, _ := json.Marshal(payload)

		sendReport(Finding{
			AssetID:     uniqueID,
			CheckType:   "Heartbeat_Metrics",
			Severity:    "Info",
			Description: string(payloadJSON),
		}, uniqueID)

		time.Sleep(time.Duration(heartbeatSec) * time.Second)
	}
}

// ──────────────────── Deep Audit ────────────────────

func runDeepAudit(agentID string) {
	fmt.Printf("[%s] Derin Denetim başlatılıyor...\n", ts())

	// 1. Yerel Admin Denetimi
	admins := collectLocalAdmins()
	if admins != nil {
		data, _ := json.Marshal(admins)
		sendReport(Finding{
			AssetID:     agentID,
			CheckType:   "Deep_Audit_Admins",
			Severity:    "Info",
			Description: string(data),
		}, agentID)
		fmt.Printf("[%s] Admin denetimi gönderildi (%d admin)\n", ts(), len(admins))
	}

	// 2. Windows Event Log Analizi
	events := collectSecurityEvents()
	if events != nil {
		data, _ := json.Marshal(events)
		sendReport(Finding{
			AssetID:     agentID,
			CheckType:   "Deep_Audit_Events",
			Severity:    "Info",
			Description: string(data),
		}, agentID)
		fmt.Printf("[%s] Güvenlik olayları gönderildi (%d olay)\n", ts(), len(events))
	}

	// 3. Dosya Bütünlüğü İzleme
	hashes := collectFileHashes()
	if hashes != nil {
		data, _ := json.Marshal(hashes)
		sendReport(Finding{
			AssetID:     agentID,
			CheckType:   "Deep_Audit_FIM",
			Severity:    "Info",
			Description: string(data),
		}, agentID)
		fmt.Printf("[%s] FIM verileri gönderildi (%d dosya)\n", ts(), len(hashes))
	}

	fmt.Printf("[%s] Derin Denetim tamamlandı.\n", ts())
}

// ── Yerel Admin Denetimi ──
func collectLocalAdmins() []string {
	cmd := exec.Command("net", "localgroup", "administrators")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("[%s] Admin listesi alınamadı: %s\n", ts(), err.Error())
		return nil
	}

	var admins []string
	lines := strings.Split(string(out), "\n")
	inMembers := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// "---" satırından sonra üye listesi başlar
		if strings.HasPrefix(line, "---") {
			inMembers = true
			continue
		}

		// "The command completed successfully" satırına kadar
		if inMembers && line != "" && !strings.HasPrefix(line, "The command completed") && !strings.HasPrefix(line, "Komut başarıyla") {
			admins = append(admins, line)
		}
	}

	return admins
}

// ── Windows Event Log Analizi ──
func collectSecurityEvents() []SecurityEventInfo {
	// Son 5 dakikadaki güvenlik olaylarını al
	// Event ID'ler: 4624 (başarılı giriş), 4625 (başarısız), 4634 (logoff), 4720 (yeni kullanıcı), 4732 (gruba ekleme)
	psScript := `
$startTime = (Get-Date).AddMinutes(-5)
$eventIDs = @(4624, 4625, 4634, 4720, 4732)
$results = @()

try {
    $events = Get-WinEvent -FilterHashtable @{
        LogName = 'Security'
        ID = $eventIDs
        StartTime = $startTime
    } -ErrorAction SilentlyContinue

    foreach ($event in $events) {
        $xml = [xml]$event.ToXml()
        $data = $xml.Event.EventData.Data

        $targetUser = ""
        $logonType = 0
        $ipAddress = ""
        $workstation = ""

        foreach ($d in $data) {
            switch ($d.Name) {
                "TargetUserName" { $targetUser = $d.'#text' }
                "LogonType" { $logonType = [int]$d.'#text' }
                "IpAddress" { $ipAddress = $d.'#text' }
                "WorkstationName" { $workstation = $d.'#text' }
            }
        }

        $results += @{
            event_id = $event.Id
            time_created = $event.TimeCreated.ToString("o")
            target_user_name = $targetUser
            logon_type = $logonType
            ip_address = $ipAddress
            workstation_name = $workstation
        }
    }
} catch {}

$results | ConvertTo-Json -Compress
`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("[%s] Security Event Log okunamadı: %s\n", ts(), err.Error())
		return nil
	}

	output := strings.TrimSpace(string(out))
	if output == "" || output == "null" {
		return []SecurityEventInfo{}
	}

	var events []SecurityEventInfo
	// PowerShell tek eleman döndürdüğünde array yapmaz, object döndürür
	if err := json.Unmarshal([]byte(output), &events); err != nil {
		// Tek eleman olabilir
		var single SecurityEventInfo
		if err2 := json.Unmarshal([]byte(output), &single); err2 == nil {
			events = append(events, single)
		}
	}

	return events
}

// ── Dosya Bütünlüğü İzleme (FIM) ──
func collectFileHashes() []FileHashInfo {
	// İzlenecek kritik dosyalar
	filePaths := []string{
		`C:\Windows\System32\drivers\etc\hosts`,
		`C:\Windows\System32\config\SAM`,
		`C:\Windows\System32\GroupPolicy\Machine\Registry.pol`,
	}

	var results []FileHashInfo

	for _, path := range filePaths {
		info, err := os.Stat(path)
		if err != nil {
			continue // Dosya yoksa atla
		}

		hash, err := computeSHA256(path)
		if err != nil {
			continue
		}

		results = append(results, FileHashInfo{
			Path:         path,
			SHA256:       hash,
			SizeBytes:    info.Size(),
			LastModified: info.ModTime().Format(time.RFC3339),
		})
	}

	return results
}

func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ──────────────────── Heartbeat ────────────────────

func buildHeartbeat() HeartbeatPayload {
	vMem, _ := mem.VirtualMemory()
	cpuP, _ := cpu.Percent(0, false)
	hInfo, _ := host.Info()

	cpuStr := "0.00%"
	if len(cpuP) > 0 {
		cpuStr = fmt.Sprintf("%.2f%%", cpuP[0])
	}

	return HeartbeatPayload{
		CPU:             cpuStr,
		RAM:             fmt.Sprintf("%.2f%%", vMem.UsedPercent),
		Uptime:          hInfo.Uptime,
		DefenderEnabled: getDefenderStatus(),
	}
}

// ──────────────────── Network ────────────────────

func sendReport(f Finding, agentID string) {
	jsonData, _ := json.Marshal(f)
	url := serverURL + reportPath

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[%s] Hata: Sunucuya ulaşılamadı.\n", ts())
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[%s] Gönderildi: %s\n", ts(), f.CheckType)

	// C2: Response'taki komutları oku ve icra et
	var response ReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return
	}

	for _, cmd := range response.Commands {
		fmt.Printf("[%s] Komut alındı: %s (ID: %s)\n", ts(), cmd.Type, cmd.ID)
		go executeCommand(cmd, agentID)
	}
}

func sendCommandResult(cmdID string, result CommandResult) {
	jsonData, _ := json.Marshal(result)
	url := fmt.Sprintf(serverURL+resultPath, cmdID)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[%s] Komut sonucu gönderilemedi: %s\n", ts(), err.Error())
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[%s] Komut sonucu gönderildi: %s → %s\n", ts(), cmdID[:8], result.Status)
}

// ──────────────────── C2 Command Executor ────────────────────

func executeCommand(cmd ServerCommand, agentID string) {
	var result CommandResult

	switch cmd.Type {
	case "screenshot":
		result = cmdScreenshot()
	case "list_processes":
		result = cmdListProcesses()
	case "sysinfo":
		result = cmdSysInfo()
	default:
		result = CommandResult{Status: "failed", Result: "Bilinmeyen komut tipi: " + cmd.Type}
	}

	sendCommandResult(cmd.ID, result)
}

// ── Screenshot ──
func cmdScreenshot() CommandResult {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return CommandResult{Status: "failed", Result: "Aktif ekran bulunamadı"}
	}

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return CommandResult{Status: "failed", Result: "Ekran yakalanamadı: " + err.Error()}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return CommandResult{Status: "failed", Result: "PNG encode hatası: " + err.Error()}
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return CommandResult{Status: "completed", Result: encoded}
}

// ── Process Listesi ──
func cmdListProcesses() CommandResult {
	procs, err := process.Processes()
	if err != nil {
		return CommandResult{Status: "failed", Result: "Süreçler alınamadı: " + err.Error()}
	}

	type ProcInfo struct {
		PID  int32   `json:"pid"`
		Name string  `json:"name"`
		CPU  float64 `json:"cpu"`
		RAM  float32 `json:"ram"`
	}

	var list []ProcInfo
	for _, p := range procs {
		name, _ := p.Name()
		cpuP, _ := p.CPUPercent()
		memP, _ := p.MemoryPercent()

		if name != "" {
			list = append(list, ProcInfo{
				PID:  p.Pid,
				Name: name,
				CPU:  cpuP,
				RAM:  memP,
			})
		}
	}

	data, _ := json.Marshal(list)
	return CommandResult{Status: "completed", Result: string(data)}
}

// ── Sistem Bilgisi ──
func cmdSysInfo() CommandResult {
	hInfo, _ := host.Info()
	vMem, _ := mem.VirtualMemory()
	cpuInfo, _ := cpu.Info()
	diskParts, _ := disk.Partitions(false)
	netIfaces, _ := net.Interfaces()

	type DiskInfo struct {
		Device     string `json:"device"`
		MountPoint string `json:"mount"`
		Total      string `json:"total"`
		Used       string `json:"used"`
		Free       string `json:"free"`
	}

	type NetInfo struct {
		Name  string   `json:"name"`
		Addrs []string `json:"addrs"`
	}

	type SysInfoPayload struct {
		OS           string     `json:"os"`
		Platform     string     `json:"platform"`
		Hostname     string     `json:"hostname"`
		KernelVer    string     `json:"kernel_version"`
		CPUModel     string     `json:"cpu_model"`
		CPUCores     int32      `json:"cpu_cores"`
		TotalRAM     string     `json:"total_ram"`
		Disks        []DiskInfo `json:"disks"`
		NetInterfaces []NetInfo `json:"net_interfaces"`
	}

	payload := SysInfoPayload{
		OS:        hInfo.OS,
		Platform:  fmt.Sprintf("%s %s", hInfo.Platform, hInfo.PlatformVersion),
		Hostname:  hInfo.Hostname,
		KernelVer: hInfo.KernelVersion,
		TotalRAM:  fmt.Sprintf("%.2f GB", float64(vMem.Total)/1024/1024/1024),
	}

	if len(cpuInfo) > 0 {
		payload.CPUModel = cpuInfo[0].ModelName
		payload.CPUCores = cpuInfo[0].Cores
	}

	for _, part := range diskParts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		payload.Disks = append(payload.Disks, DiskInfo{
			Device:     part.Device,
			MountPoint: part.Mountpoint,
			Total:      fmt.Sprintf("%.2f GB", float64(usage.Total)/1024/1024/1024),
			Used:       fmt.Sprintf("%.2f GB", float64(usage.Used)/1024/1024/1024),
			Free:       fmt.Sprintf("%.2f GB", float64(usage.Free)/1024/1024/1024),
		})
	}

	for _, iface := range netIfaces {
		var addrs []string
		for _, a := range iface.Addrs {
			addrs = append(addrs, a.Addr)
		}
		if len(addrs) > 0 {
			payload.NetInterfaces = append(payload.NetInterfaces, NetInfo{
				Name:  iface.Name,
				Addrs: addrs,
			})
		}
	}

	data, _ := json.Marshal(payload)
	return CommandResult{Status: "completed", Result: string(data)}
}

// ──────────────────── Collectors ────────────────────

func getInstalledSoftware() []string {
	var softwareList []string
	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		names, _ := k.ReadSubKeyNames(-1)
		for _, name := range names {
			subKey, err := registry.OpenKey(registry.LOCAL_MACHINE, path+`\`+name, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			displayName, _, err := subKey.GetStringValue("DisplayName")
			if err == nil && displayName != "" {
				softwareList = append(softwareList, displayName)
			}
			subKey.Close()
		}
		k.Close()
	}
	return softwareList
}

func getDefenderStatus() bool {
	cmd := exec.Command("powershell", "-Command", "(Get-MpComputerStatus).AMServiceEnabled")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "True"
}

// ──────────────────── Helpers ────────────────────

func ts() string {
	return time.Now().Format("15:04:05")
}

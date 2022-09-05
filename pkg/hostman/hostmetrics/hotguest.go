package hostmetrics

import (
	"fmt"
	"time"
	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
	modules "yunion.io/x/onecloud/pkg/mcclient/modules/compute"
	"yunion.io/x/onecloud/pkg/util/influxdb"
)

const (
	VM_CPU     = "vm_cpu"
	VM_MEM     = "vm_mem"
	AGENT_DISK = "agent_disk"
)

type Usage struct {
	CpuUsage  string `json:"cpu_usage"`
	MemUsage  string `json:"mem_usage"`
	DiskUsage string `json:"disk_usage"`
}

func GetInfluxAgentData(dbname []string, vm_id string) map[string]Usage {
	vm_data := map[string]Usage{}
	usgae_type := AnalysisData(dbname, vm_id)
	vm_data[vm_id] = usgae_type
	return vm_data
}

func AnalysisData(dbname []string, vm_id string) Usage {
	db := influxdb.NewInfluxdbWithDebug(InfluxdbReader, true)
	usage_type := Usage{}
	for _, name := range dbname {
		query := fmt.Sprintf("SELECT * FROM %s where vm_id = '%s' order by time desc limit 1", name, vm_id)
		rtn, err := db.Query(query)
		if err != nil {
			log.Errorf("query vm err=%v", err)
		}
		if len(rtn) <= 0 {
			break
		}
		for _, result := range rtn {
			for _, obj := range result {
				var usage = "0"
				for i, col := range obj.Columns {
					if len(obj.Values[0]) <= 0 {
						continue
					}
					switch name {
					case AGENT_DISK:
						if col == "used_percent" {
							if val := obj.Values[0][i]; val != nil {
								usage, _ = val.GetString()
								usage_type.DiskUsage = usage
							}
						}
					case VM_MEM:
						if col == "used_percent" {
							if val := obj.Values[0][i]; val != nil {
								usage, _ = val.GetString()
								usage_type.MemUsage = usage
							}
						}
					case VM_CPU:
						if col == "usage_active" {
							if val := obj.Values[0][i]; val != nil {
								usage, _ = val.GetString()
								usage_type.CpuUsage = usage
							}
						}
					}
				}
			}
		}
	}
	return usage_type
}

func SyncHostData(reportData map[string]Usage) {
	adminToken := auth.AdminCredential()
	regions := adminToken.GetRegions()
	s := auth.GetAdminSession(nil, regions[0])
	body := jsonutils.NewDict()
	for key, value := range reportData {
		body.Set("cpu_usage", jsonutils.NewString(value.CpuUsage))
		body.Set("mem_usage", jsonutils.NewString(value.MemUsage))
		body.Set("disk_usage", jsonutils.NewString(value.DiskUsage))
		if _, err := modules.Servers.Update(s, key, body); err != nil {
			log.Errorf("sync update host data err=[%v]", err)
			return
		}
	}
}

func UpdateDB(vm_id, usage, vm_name string) {
	adminToken := auth.AdminCredential()
	regions := adminToken.GetRegions()
	s := auth.GetAdminSession(nil, regions[0])
	body := jsonutils.NewDict()
	if vm_name == "vm_cpu" {
		body.Set("cpu_usage", jsonutils.NewString(usage))
	}

	if vm_name == "vm_mem" {
		body.Set("mem_usage", jsonutils.NewString(usage))
	}

	if _, err := modules.Servers.Update(s, vm_id, body); err != nil {
		log.Errorf("sync update host data err=[%v]", err)
	}
}

func GetServerId() []string {
	filter := jsonutils.NewDict()
	adminToken := auth.AdminCredential()
	regions := adminToken.GetRegions()
	s := auth.GetAdminSession(nil, regions[0])
	results, _ := modules.Servers.List(s, filter)
	vm_ids := []string{}
	for _, result := range results.Data {
		id, err := result.GetString("id")
		if id != "" {
			vm_ids = append(vm_ids, id)
		}
		if err != nil {
			log.Infof("error")
		}
	}
	return vm_ids
}

func HostGuestRun() {
	dbNames := []string{AGENT_DISK}
	result_list := GetServerId()
	for _, vm_id := range result_list {
		if vm_data := GetInfluxAgentData(dbNames, vm_id); len(vm_data) > 0 {
			SyncHostData(vm_data)
		}
	}
}

func TickRun() {
	done := make(chan bool)
	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			HostGuestRun()
		}
	}
}

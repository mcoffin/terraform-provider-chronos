package chronos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

type job struct {
	Schedule string `json:"schedule"`
	Name string `json:"name"`
	Cpus float64 `json:"cpus"`
	Mem float64 `json:"mem"`
	Disk float64 `json:"disk"`
	Uris []string `json:"uris"`
	Container container `json:"container"`
	Command string `json:"command"`
	Env []envEntry `json:"environmentVariables"`
}

type envEntry struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type container struct {
	Type string `json:"type"`
	Image string `json:"image"`
	Network string `json:"network"`
	ForcePullImage bool `json:"forcePullImage"`
	Volumes []volume `json:"volumes"`
}

type volume struct {
	ContainerPath string `json:"containerPath"`
	HostPath string `json:"hostPath"`
	Mode string `json:"mode"`
}

func resourceChronosJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceChronosJobCreate,
		Read: resourceChronosJobRead,
		Update: resourceChronosJobCreate,
		Delete: resourceChronosJobDelete,
		Schema: map[string]*schema.Schema{
			"schedule": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"job_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"cpus": &schema.Schema{
				Type: schema.TypeFloat,
				Required: true,
				ForceNew: false,
			},
			"mem": &schema.Schema{
				Type: schema.TypeFloat,
				Required: true,
				ForceNew: false,
			},
			"disk": &schema.Schema{
				Type: schema.TypeFloat,
				Required: true,
				ForceNew: false,
			},
			"uris": &schema.Schema{
				Type: schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"container": &schema.Schema{
				Type: schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
						"image": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
						"network": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Default: "BRIDGE",
						},
						"force_pull_image": &schema.Schema{
							Type: schema.TypeBool,
							Optional: true,
							Default: false,
						},
						"volume": &schema.Schema{
							Type: schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_path": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
									"host_path": &schema.Schema{
										Type: schema.TypeString,
										Required: true,
									},
									"mode": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Default: "RW",
									},
								},
							},
						},
					},
				},
			},
			"command": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"env": &schema.Schema{
				Type: schema.TypeMap,
				Required: true,
				ForceNew: false,
			},
		},
	}
}

func jobFromResource(rd *schema.ResourceData) *job {
	j := new(job)

	if v, ok := rd.GetOk("schedule"); ok {
		j.Schedule = v.(string)
	}

	if v, ok := rd.GetOk("job_id"); ok {
		j.Name = v.(string)
	}

	if v, ok := rd.GetOk("cpus"); ok {
		j.Cpus = v.(float64)
	}

	if v, ok := rd.GetOk("mem"); ok {
		j.Mem = v.(float64)
	}

	if v, ok := rd.GetOk("disk"); ok {
		j.Disk = v.(float64)
	}

	if v, ok := rd.GetOk("command"); ok {
		j.Command = v.(string)
	}

	if v, ok := rd.GetOk("uris.#"); ok {
		uris := make([]string, v.(int))
		for i, _ := range uris {
			uris[i] = rd.Get(fmt.Sprintf("uris.%d", i)).(string)
		}
		j.Uris = uris
	}

	if v, ok := rd.GetOk("env"); ok {
		envMap := v.(map[string]interface{})
		env := make([]envEntry, len(envMap))
		i := 0
		for k, v := range envMap {
			e := envEntry{
				Name: k,
				Value: v.(string),
			}

			env[i] = e
			i++
		}
		j.Env = env
	}

	c := new(container)

	if v, ok := rd.GetOk("container.0.type"); ok {
		c.Type = v.(string)
	}

	if v, ok := rd.GetOk("container.0.image"); ok {
		c.Image = v.(string)
	}

	if v, ok := rd.GetOk("container.0.network"); ok {
		c.Network = v.(string)
	}

	if v, ok := rd.GetOk("container.0.force_pull_image"); ok {
		c.ForcePullImage = v.(bool)
	}

	if v, ok := rd.GetOk("container.0.volume.#"); ok {
		volumes := make([]volume, v.(int))

		for i, _ := range volumes {
			vol := volume{
				ContainerPath: rd.Get(fmt.Sprintf("container.0.volume.%d.container_path")).(string),
				HostPath: rd.Get(fmt.Sprintf("container.0.volume.%d.host_path")).(string),
				Mode: rd.Get(fmt.Sprintf("container.0.volume.%d.mode")).(string),
			}
			volumes[i] = vol
		}

		c.Volumes = volumes
	}

	j.Container = *c

	return j
}

func setAndPartial(rd *schema.ResourceData, key string, value interface{}) {
	rd.Set(key, value)
	rd.SetPartial(key)
}

func setSchemaFieldsForJob(j *job, rd *schema.ResourceData) {
	setAndPartial(rd, "schedule", j.Schedule)
	setAndPartial(rd, "job_id", j.Name)
	setAndPartial(rd, "cpus", j.Cpus)
	setAndPartial(rd, "mem", j.Mem)
	setAndPartial(rd, "disk", j.Disk)
	setAndPartial(rd, "uris", j.Uris)
	setAndPartial(rd, "command", j.Command)

	envMap := make(map[string]interface{})
	for _, e := range j.Env {
		envMap[e.Name] = e.Value
	}
	setAndPartial(rd, "env", envMap)

	containerMap := make(map[string]interface{})
	containerMap["type"] = j.Container.Type
	containerMap["image"] = j.Container.Image
	containerMap["network"] = j.Container.Network
	containerMap["force_pull_image"] = j.Container.ForcePullImage

	volumeMaps := make([]map[string]interface{}, len(j.Container.Volumes))
	for idx, volume := range j.Container.Volumes {
		volumeMap := make(map[string]interface{})
		volumeMap["container_path"] = volume.ContainerPath
		volumeMap["host_path"] = volume.HostPath
		volumeMap["mode"] = volume.Mode
		volumeMaps[idx] = volumeMap
	}
	containerMap["volume"] = volumeMaps

	setAndPartial(rd, "container", containerMap)
}

func checkStatusCodeIsSuccessful(statusCode int) error {
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("Received bad status code %d", statusCode)
	}
	return nil
}

func resourceChronosJobCreate(rd *schema.ResourceData, meta interface{}) error {
	chronosConfig := meta.(Config)

	j := jobFromResource(rd)

	jsonBytes, err := json.Marshal(j)
	if err != nil {
		return err
	}

	req, err := chronosConfig.CreateRequest("POST", chronosConfig.GetCreateUrl(), "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if err = checkStatusCodeIsSuccessful(resp.StatusCode); err != nil {
		return err
	}

	rd.Partial(true)
	rd.SetId(j.Name)
	setSchemaFieldsForJob(j, rd)
	rd.Partial(false)

	return resourceChronosJobRead(rd, meta)
}

func resourceChronosJobRead(rd *schema.ResourceData, meta interface{}) error {
	chronosConfig := meta.(Config)

	j, err := chronosConfig.getJob(rd.Id())

	if err != nil {
		if err == ErrJobDoesNotExist {
			rd.SetId("")
			return nil
		}
		return err
	}

	if j.Name == "" {
		rd.SetId("")
	}

	setSchemaFieldsForJob(j, rd)
	return nil
}

func resourceChronosJobDelete(rd *schema.ResourceData, meta interface{}) error {
	chronosConfig := meta.(Config)

	j := jobFromResource(rd)

	req, err := chronosConfig.CreateRequest("DELETE", fmt.Sprintf("%s/scheduler/job/%s", chronosConfig.Url.String(), j.Name), "", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return checkStatusCodeIsSuccessful(resp.StatusCode)
}

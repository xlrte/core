package gcp

import (
	_ "embed"
	"fmt"

	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

//go:embed templates/cloudsql.tf
var cloudSqlMain string

type cloudSql struct {
	baseDir                    string
	DbName                     string `yaml:"name"`
	DBType                     string `yaml:"type"`
	MachineType                string `yaml:"machine_type"`
	Size                       int    `yaml:"storage_size"`
	DeleteProtection           bool   `yaml:"delete_protection"`
	MaintenanceWindowHour      int    `yaml:"maintenance_window_hour"`
	MaintenanceWindowDay       int    `yaml:"maintenance_window_day"`
	PointInTimeRecoveryEnabled bool   `yaml:"point_in_time_recovery_enabled"`
	BackupEnabled              *bool  `yaml:"backup_enabled"`
	BackupStartTime            string `yaml:"backup_start_time"`
	NetworkLink                string `yaml:"-"`
}

func (rt *cloudSql) Load(d *api.ResourceDefinition) ([]api.Resource, []api.DependencyBinding, error) {
	var dbs []*cloudSql
	var resources []cloudSql
	var rs []api.Resource
	var bindings []api.DependencyBinding

	if d.Name == "cloudsql" {
		err := yaml.Unmarshal(d.ServiceConfig, &dbs)
		if err != nil {
			return nil, nil, err
		}
		if d.ResourceConfig != nil {
			err = yaml.Unmarshal(*d.ResourceConfig, &resources)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	network := &privateNetwork{baseDir: rt.baseDir, MinInstances: 2, MaxInstances: 3, InstanceType: "f1-micro"}
	err := d.GetConfig("vpc_access_connector", network)
	if err != nil {
		return nil, nil, err
	}
	rs = append(rs, network)
	bindings = append(bindings, api.DependencyBinding{
		DependedOnBy: d.DependedOnBy,
		Privileges:   api.Owner,
		Identity:     network.Identity(),
		Config:       network.configurator(),
	})

	for _, db := range dbs {
		for _, r := range resources {
			if db.DbName == r.DbName {
				db.Size = r.Size
				db.MachineType = r.MachineType
				db.DeleteProtection = r.DeleteProtection
				db.MaintenanceWindowDay = r.MaintenanceWindowDay
				db.MaintenanceWindowHour = r.MaintenanceWindowHour
				db.PointInTimeRecoveryEnabled = r.PointInTimeRecoveryEnabled
				db.BackupEnabled = r.BackupEnabled
				db.BackupStartTime = r.BackupStartTime
				break
			}
		}
		if db.Size == 0 {
			db.Size = 10
		}
		if db.MachineType == "" {
			db.MachineType = "db-f1-micro"
		}
		if db.DBType == "postgres" {
			db.DBType = "POSTGRES_13"
		} else {
			return nil, nil, fmt.Errorf("at the moment 'postgres' is the only supported database")
		}
		if db.MaintenanceWindowDay == 0 {
			db.MaintenanceWindowDay = 7
		}
		if db.MaintenanceWindowHour == 0 {
			db.MaintenanceWindowHour = 7
		}
		if db.BackupEnabled == nil {
			soTrue := true
			db.BackupEnabled = &soTrue
		}
		if db.BackupStartTime == "" {
			db.BackupStartTime = "04:00"
		}
		db.baseDir = rt.baseDir
		identity := api.ResourceIdentity{Type: "cloudsql", ID: db.DbName}
		passwordRef := api.SecretRef{
			Name: "PASSWORD",
			Type: api.RandomString,
		}
		userRef := api.SecretRef{
			Name: "USER",
			Type: api.RandomString,
		}
		// DB dependency of Service
		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   api.Owner,
			Identity:     identity,
			Config:       db,
			SecretRefs:   []api.SecretRef{passwordRef, userRef},
		})

		// Network dependency of DB
		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: identity,
			Privileges:   api.Owner,
			Identity:     network.Identity(),
			Config:       network.configurator(),
		})
		rs = append(rs, db)
	}
	return rs, bindings, nil
}

func (r *cloudSql) Configure() error {
	return applyTerraformTemplates(r.baseDir, []crFile{
		{"cloudsql.tf", cloudSqlMain},
	}, r)
}

func (r *cloudSql) Name() string {
	return "cloudsql"
}

func (r *cloudSql) Identity() api.ResourceIdentity {
	return api.ResourceIdentity{Type: "cloudsql", ID: r.DbName}
}

func (r *cloudSql) ConfigureResource(resource api.Resource) error {
	cloudrun, ok := resource.(*cloudRunConfig)
	if ok {
		host := fmt.Sprintf("module.%s-%s.master_private_ip", "cloudsql", r.DbName)
		dependency := fmt.Sprintf("module.%s-%s", "cloudsql", r.DbName)
		if cloudrun.Env.Refs == nil {
			cloudrun.Env.Refs = make(map[string]string)
		}
		if cloudrun.Env.Secrets == nil {
			cloudrun.Env.Secrets = make(map[string]string)
		}
		pwdSecret := fmt.Sprintf("secret-%s_PASSWORD", r.Identity().String())
		userSecret := fmt.Sprintf("secret-%s_USER", r.Identity().String())
		pwdSecretId := fmt.Sprintf("secret-%s_PASSWORD.secret_id", r.Identity().String())
		userSecretId := fmt.Sprintf("secret-%s_USER.secret_id", r.Identity().String())
		cloudrun.Env.Refs[fmt.Sprintf("DB_%s_HOST", r.DbName)] = host
		cloudrun.DependsOn = append(cloudrun.DependsOn, dependency)
		cloudrun.DependsOn = append(cloudrun.DependsOn, toDependency(pwdSecret))
		cloudrun.DependsOn = append(cloudrun.DependsOn, toDependency(userSecret))
		cloudrun.Env.Secrets[fmt.Sprintf("DB_%s_PASSWORD", r.DbName)] = toDependency(pwdSecretId)
		cloudrun.Env.Secrets[fmt.Sprintf("DB_%s_USER", r.DbName)] = toDependency(userSecretId)
	}
	return nil
}

func toDependency(name string) string {
	return fmt.Sprintf("module.%s", name)
}

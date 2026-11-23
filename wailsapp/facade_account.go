package main

import (
	"errors"

	"cloudshelf/internal/awss3"
	"cloudshelf/internal/model"
	"cloudshelf/internal/rws"
	"cloudshelf/internal/sharepoint"
)

// AccountDTO 是前后端之间的账号传输结构，字段全为可序列化基本类型。
// 用它隔离 model.Account 的 time.Time 与仅内存使用的字段，避免绑定生成的不确定性；
// 与 model.Account 的双向转换只在 facade 内完成，后端包零改动。
type AccountDTO struct {
	ID              int64
	Name            string
	Provider        string
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UsePathStyle    bool
	TenantID        string
	ClientID        string
	ClientSecret    string
	SiteURL         string
}

func accountToDTO(a model.Account) AccountDTO {
	return AccountDTO{
		ID:              a.ID,
		Name:            a.Name,
		Provider:        a.Provider,
		Endpoint:        a.Endpoint,
		Region:          a.Region,
		AccessKeyID:     a.AccessKeyID,
		SecretAccessKey: a.SecretAccessKey,
		Bucket:          a.Bucket,
		UsePathStyle:    a.UsePathStyle,
		TenantID:        a.TenantID,
		ClientID:        a.ClientID,
		ClientSecret:    a.ClientSecret,
		SiteURL:         a.SiteURL,
	}
}

func (d AccountDTO) toModel() model.Account {
	return model.Account{
		ID:              d.ID,
		Name:            d.Name,
		Provider:        d.Provider,
		Endpoint:        d.Endpoint,
		Region:          d.Region,
		AccessKeyID:     d.AccessKeyID,
		SecretAccessKey: d.SecretAccessKey,
		Bucket:          d.Bucket,
		UsePathStyle:    d.UsePathStyle,
		TenantID:        d.TenantID,
		ClientID:        d.ClientID,
		ClientSecret:    d.ClientSecret,
		SiteURL:         d.SiteURL,
	}
}

// ---- 账号 CRUD ----

func (a *App) ListAccounts() ([]AccountDTO, error) {
	s := a.currentStore()
	if s == nil {
		return nil, errors.New("账号库未解锁")
	}
	list, err := s.ListAccounts()
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0, len(list))
	for _, acc := range list {
		out = append(out, accountToDTO(acc))
	}
	return out, nil
}

func (a *App) AddAccount(acc AccountDTO) error {
	s := a.currentStore()
	if s == nil {
		return errors.New("账号库未解锁")
	}
	m := acc.toModel()
	return s.AddAccount(&m)
}

func (a *App) UpdateAccount(acc AccountDTO) error {
	s := a.currentStore()
	if s == nil {
		return errors.New("账号库未解锁")
	}
	m := acc.toModel()
	return s.UpdateAccount(&m)
}

func (a *App) DeleteAccount(id int64) error {
	s := a.currentStore()
	if s == nil {
		return errors.New("账号库未解锁")
	}
	return s.DeleteAccount(id)
}

// TestAccount 按 Provider 分流测试连接（旧版 doTest 的逻辑）。
func (a *App) TestAccount(acc AccountDTO) error {
	m := acc.toModel()
	if m.IsSharePoint() {
		return sharepoint.TestConnection(spConfig(&m))
	}
	if m.IsRWS() {
		return rws.TestConnection(&m)
	}
	return awss3.TestConnection(&m)
}

// spConfig 复刻 internal/ui/ui.go:1706 的账号 → SharePoint 配置映射，
// 供 TestAccount 及后续阶段的 SharePoint 浏览复用。
func spConfig(acc *model.Account) *sharepoint.Config {
	return &sharepoint.Config{
		TenantID:     acc.TenantID,
		ClientID:     acc.ClientID,
		ClientSecret: acc.ClientSecret,
		SiteURL:      acc.SiteURL,
		ReadOnly:     acc.ReadOnly,
	}
}

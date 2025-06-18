package datasrcservice

import (
	"gorm.io/gorm"
	"saurfang/internal/models/datasource"
	"saurfang/internal/repository/base"
)

// DataSourceService
type DataSourceService struct {
	base.BaseGormRepository[datasource.SaurfangDatasources]
}

// NewDataSourceService
func NewDataSourceService(db *gorm.DB) *DataSourceService {
	return &DataSourceService{
		BaseGormRepository: base.BaseGormRepository[datasource.SaurfangDatasources]{DB: db},
	}
}
func (d *DataSourceService) Service_CreateDataSource(ds *datasource.SaurfangDatasources) error {
	return d.DB.Create(&ds).Error
}
func (d *DataSourceService) Service_DeleteDataSource(id uint) error {
	return d.DB.Delete(&datasource.SaurfangDatasources{}, "id = ?", id).Error
}
func (d *DataSourceService) Service_UpdateDataSource(ds *datasource.SaurfangDatasources) error {
	return d.DB.Model(ds).Where("id = ?", ds.ID).Updates(ds).Error
}
func (d *DataSourceService) Service_ShowDataSource() (*[]datasource.SaurfangDatasources, error) {
	var source []datasource.SaurfangDatasources
	if err := d.DB.Find(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}
func (d *DataSourceService) Service_ShowDataSourceByID(id uint) (*datasource.SaurfangDatasources, error) {
	var source datasource.SaurfangDatasources
	if err := d.DB.Where("id = ?", id).First(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

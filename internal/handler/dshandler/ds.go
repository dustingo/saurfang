package dshandler

import (
	"github.com/gofiber/fiber/v3"
	"saurfang/internal/models/amis"
	"saurfang/internal/models/datasource"
	"saurfang/internal/service/datasrcservice"
	"strconv"
)

// DataSourceHandler
type DataSourceHandler struct {
	datasrcservice.DataSourceService
}

func NewSDataSourceHandler(svc *datasrcservice.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{*svc}
}

// Handler_CreateDS 创建数据源记录
func (d *DataSourceHandler) Handler_CreateDS(c fiber.Ctx) error {
	var ds datasource.SaurfangDatasources
	if err := c.Bind().Body(&ds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := d.DataSourceService.Service_CreateDataSource(&ds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_DeleteDS 删除数据源记录
func (d *DataSourceHandler) Handler_DeleteDS(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := d.DataSourceService.Service_DeleteDataSource(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_UpdateDS 更新数据源记录
func (d *DataSourceHandler) Handler_UpdateDS(c fiber.Ctx) error {
	var ds datasource.SaurfangDatasources
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := c.Bind().Body(&ds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if _, err := d.DataSourceService.Service_ShowDataSourceByID(uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	if err := d.DataSourceService.Service_UpdateDataSource(&ds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
	})
}

// Handler_ShowDS 展示数据源记录
func (d *DataSourceHandler) Handler_ShowDS(c fiber.Ctx) error {
	ds, err := d.DataSourceService.Service_ShowDataSource()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    ds,
	})
}

func (d *DataSourceHandler) Handler_ShowDSByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	ds, err := d.DataSourceService.Service_ShowDataSourceByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data":    ds,
	})
}
func (d *DataSourceHandler) Handler_SelectDS(c fiber.Ctx) error {
	ds, err := d.DataSourceService.Service_ShowDataSource()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  1,
			"message": err.Error(),
		})
	}
	var ops []amis.AmisOptions
	var op amis.AmisOptions
	for _, sn := range *ds {
		op.Label = sn.Label
		op.Value = int(sn.ID)
		ops = append(ops, op)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  0,
		"message": "success",
		"data": fiber.Map{
			"options": ops,
		},
	})
}

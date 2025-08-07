package dshandler

import (
	"saurfang/internal/models/amis"
	"saurfang/internal/models/datasource"
	"saurfang/internal/repository/base"
	"saurfang/internal/tools/pkg"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// DataSourceHandler
type DataSourceHandler struct {
	base.BaseGormRepository[datasource.Datasources]
}

// Handler_CreateDS 创建数据源记录
func (d *DataSourceHandler) Handler_CreateDS(c fiber.Ctx) error {
	var ds datasource.Datasources
	if err := c.Bind().Body(&ds); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := d.Create(&ds); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to create ds", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_DeleteDS 删除数据源记录
func (d *DataSourceHandler) Handler_DeleteDS(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := d.Delete(uint(id)); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to delete ds", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_UpdateDS 更新数据源记录
func (d *DataSourceHandler) Handler_UpdateDS(c fiber.Ctx) error {
	var ds datasource.Datasources
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	if err := c.Bind().Body(&ds); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	ds.ID = uint(id)
	if _, err := d.ListByID(ds.ID); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "failed to update ds", err.Error(), fiber.Map{})
	}
	if err := d.Update(uint(id), &ds); err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "failed to update ds", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", nil)
}

// Handler_ShowDS 展示数据源记录
func (d *DataSourceHandler) Handler_ShowDS(c fiber.Ctx) error {
	ds, err := d.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "request error", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", ds)
}

func (d *DataSourceHandler) Handler_ShowDSByID(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusBadRequest, 1, "request error", err.Error(), fiber.Map{})
	}
	ds, err := d.ListByID(uint(id))
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "request error", err.Error(), fiber.Map{})
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", ds)
}
func (d *DataSourceHandler) Handler_SelectDS(c fiber.Ctx) error {
	ds, err := d.List()
	if err != nil {
		return pkg.NewAppResponse(c, fiber.StatusInternalServerError, 1, "request error", err.Error(), fiber.Map{})
	}
	var ops []amis.AmisOptions
	var op amis.AmisOptions
	for _, sn := range ds {
		op.Label = sn.Label
		op.Value = int(sn.ID)
		ops = append(ops, op)
	}
	return pkg.NewAppResponse(c, fiber.StatusOK, 0, "success", "", fiber.Map{
		"options": ops,
	})
}

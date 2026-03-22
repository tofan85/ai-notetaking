package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/interfaces"
	"ai-notetaking-be/internal/pkg/serverutils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.elastic.co/apm"
)

type notebookController struct {
	service interfaces.INotebookService
}

func NewNotebookController(service interfaces.INotebookService) interfaces.INotebookController {
	return &notebookController{service: service}
}

func (c *notebookController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/notebook/v1")
	h.Post("", c.Create)
	h.Get(":id", c.Show)
	h.Put(":id", c.Update)
	h.Delete(":id", c.Delete)
	h.Put(":id/movenotebook", c.MoveNotebook)
	h.Get("", c.GetAllRoutes)
}

func (c *notebookController) GetAllRoutes(ctx *fiber.Ctx) error {
	span, spanTx := apm.StartSpan(ctx.Context(), "GetAll", "Controller")
	defer span.End()

	res, err := c.service.GetAll(spanTx)

	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success Get List All", res))

}
func (c *notebookController) Create(ctx *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(ctx.Context(), "Register", "Controller")
	defer span.End()
	var req dto.CreateNotebookRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.CreateNotebook(spanCtx, &req)
	if err != nil {
		return err
	}
	return ctx.JSON(serverutils.SuccessResponse("Success create notebook", res))
}

func (c *notebookController) Show(ctx *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(ctx.Context(), "Register", "Controller")
	defer span.End()
	idparam := ctx.Params("id")
	id, _ := uuid.Parse(idparam)
	res, err := c.service.Show(spanCtx, id)
	if err != nil {
		return err
	}
	return ctx.JSON(serverutils.SuccessResponse("Success get notebook", res))

}

func (c *notebookController) Update(ctx *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(ctx.Context(), "Register", "Controller")
	defer span.End()
	idparam := ctx.Params("id")
	id, _ := uuid.Parse(idparam)
	var req dto.UpdateNotebookRequest
	req.ID = id
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	res, err := c.service.UpdateNotebook(spanCtx, &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success update notebook", res))
}

func (c *notebookController) Delete(ctx *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(ctx.Context(), "Register", "Controller")
	defer span.End()
	idparam := ctx.Params("id")
	id, _ := uuid.Parse(idparam)

	err := c.service.Delete(spanCtx, id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse[any]("Success delete notebook", nil))
}

func (c *notebookController) MoveNotebook(ctx *fiber.Ctx) error {
	var req dto.MoveNotebookRequest
	span, spanCtx := apm.StartSpan(ctx.Context(), "MoveNotebook", "Repository")
	defer span.End()

	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	req.ID = id

	res, err := c.service.MoveNotebook(spanCtx, &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Succes move notebook", res))

}

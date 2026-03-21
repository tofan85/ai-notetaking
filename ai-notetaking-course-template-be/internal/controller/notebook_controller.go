package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/interfaces"
	"ai-notetaking-be/internal/pkg/serverutils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
}

func (c *notebookController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateNotebookRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.CreateNotebook(ctx.Context(), &req)
	if err != nil {
		return err
	}
	return ctx.JSON(serverutils.SuccessResponse("Success create notebook", res))
}

func (c *notebookController) Show(ctx *fiber.Ctx) error {
	idparam := ctx.Params("id")
	id, _ := uuid.Parse(idparam)
	res, err := c.service.Show(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(serverutils.SuccessResponse("Success get notebook", res))

}

func (c *notebookController) Update(ctx *fiber.Ctx) error {
	idparam := ctx.Params("id")
	id, _ := uuid.Parse(idparam)
	var req dto.UpdateNotebookRequest
	req.ID = id
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	res, err := c.service.UpdateNotebook(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success update notebook", res))
}

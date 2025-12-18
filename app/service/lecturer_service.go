package service

import (
	"gouas/app/repository"
	"gouas/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LecturerService interface {
	GetAll(c *fiber.Ctx) error
	GetAdvisees(c *fiber.Ctx) error
}

type lecturerService struct {
	repo repository.LecturerRepository
}

func NewLecturerService(repo repository.LecturerRepository) LecturerService {
	return &lecturerService{repo}
}

func (s *lecturerService) GetAll(c *fiber.Ctx) error {
	lecturers, err := s.repo.FindAll()
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Lecturer list", lecturers))
}

func (s *lecturerService) GetAdvisees(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	students, err := s.repo.FindAdvisees(id)
	if err != nil {
		return c.Status(500).JSON(helper.APIResponse("error", err.Error(), nil))
	}
	return c.Status(200).JSON(helper.APIResponse("success", "Advisees list", students))
}
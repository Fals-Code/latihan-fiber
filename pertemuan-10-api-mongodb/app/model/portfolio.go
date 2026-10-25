package model

import "time"

// Portfolio is the API representation of one student's portfolio.
// MongoDB-specific identifiers are represented as strings at this layer.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
}

// Project is an embedded portfolio item returned by the API.
type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}

// Certificate is an embedded certificate returned by the API.
type Certificate struct {
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

// PortfolioRequest replaces all editable portfolio fields, including both arrays.
type PortfolioRequest struct {
	Headline     string               `json:"headline" validate:"required"`
	Bio          string               `json:"bio" validate:"required"`
	Projects     []ProjectRequest     `json:"projects" validate:"max=20,dive"`
	Certificates []CertificateRequest `json:"certificates" validate:"dive"`
}

// ProjectRequest contains fields accepted when adding or replacing a project.
// The server assigns the project ID so clients cannot choose database identifiers.
type ProjectRequest struct {
	Title        string   `json:"title" validate:"required"`
	Technologies []string `json:"technologies" validate:"required,min=1,dive,required"`
	Link         string   `json:"link" validate:"required,url"`
	Year         int      `json:"year" validate:"required,gt=0"`
}

// CertificateRequest contains fields accepted when replacing a portfolio.
type CertificateRequest struct {
	Name   string    `json:"name" validate:"required"`
	Issuer string    `json:"issuer" validate:"required"`
	Date   time.Time `json:"date" validate:"required"`
}

// AddProjectRequest is the body for the endpoint that appends one project.
type AddProjectRequest = ProjectRequest

// TechnologyCount is one row in the portfolio technology statistics response.
type TechnologyCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}

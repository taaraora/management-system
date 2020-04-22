package postgres

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/omc-college/management-system/pkg/rbac/models"
)

type RolesRepository struct {
	DB *sql.DB
}

func NewDB(dsn string) (*RolesRepository, error) {
	db, err := sql.Open("pgx", dsn)

	return &RolesRepository{
		DB: db,
	}, err
}

func GetAllRoles(repository *RolesRepository) (map[int]*models.Role, error) {
	query := `SELECT roles.id, roles.name, features.id, features.name, endpoints.id, endpoints.path, endpoints.method
			  FROM roles LEFT JOIN "rolesToFeatures"
			  ON roles.id = "rolesToFeatures"."roleId"
			  LEFT JOIN features
			  ON "rolesToFeatures"."featureId" = features.id
			  LEFT JOIN "featuresToEndpoints"
			  ON features.id = "featuresToEndpoints"."featureId"
			  LEFT JOIN endpoints
			  ON "featuresToEndpoints"."endpointId" = endpoints.id`

	roles := make(map[int]*models.Role)

	rows, err := repository.DB.Query(query)
	if err != nil {
		return map[int]*models.Role{}, QueryError{queryErrorMessage, err}
	}

	for rows.Next() {
		var role models.Role
		var feature models.FeatureEntry
		var endpoint models.Endpoint
		isRoleExisting := false
		isFeatureExisting := false

		err := rows.Scan(&role.ID, &role.Name, &feature.ID, &feature.Name, &endpoint.ID, &endpoint.Path, &endpoint.Method)
		if err != nil {
			return map[int]*models.Role{}, ScanError{scanErrorMessage, err}
		}

		_, isRoleExisting = roles[role.ID]
		if !isRoleExisting {
			role.Entries = make(map[int]*models.FeatureEntry)
			roles[role.ID] = &role
		}

		_, isFeatureExisting = roles[role.ID].Entries[feature.ID]
		if !isFeatureExisting {
			feature.Endpoints = make(map[int]*models.Endpoint)
			roles[role.ID].Entries[feature.ID] = &feature
		}

		roles[role.ID].Entries[feature.ID].Endpoints[endpoint.ID] = &endpoint
	}

	err = rows.Err()
	if err != nil {
		return map[int]*models.Role{}, ScanError{scanErrorMessage, err}
	}

	return roles, nil
}

func GetRole(repository *RolesRepository, id int) (*models.Role, error) {
	query := `SELECT roles.id, roles.name, features.id, features.name, endpoints.id, endpoints.path, endpoints.method
			  FROM roles LEFT JOIN "rolesToFeatures"
			  ON roles.id = "rolesToFeatures"."roleId"
			  LEFT JOIN features
			  ON "rolesToFeatures"."featureId" = features.id
			  LEFT JOIN "featuresToEndpoints"
			  ON features.id = "featuresToEndpoints"."featureId"
			  LEFT JOIN endpoints
			  ON "featuresToEndpoints"."endpointId" = endpoints.id
			  WHERE roles.id = $1`

	var role models.Role
	role.Entries = make(map[int]*models.FeatureEntry)

	rows, err := repository.DB.Query(query, id)
	if err != nil {
		return &models.Role{}, QueryError{queryErrorMessage, err}
	}

	for rows.Next() {
		var feature models.FeatureEntry
		var endpoint models.Endpoint
		isFeatureExisting := false

		err = rows.Scan(&role.ID, &role.Name, &feature.ID, &feature.Name, &endpoint.ID, &endpoint.Path, &endpoint.Method)
		if err != nil {
			return &models.Role{}, ScanError{scanErrorMessage, err}
		}

		isFeatureExisting = false

		_, isFeatureExisting = role.Entries[feature.ID]
		if !isFeatureExisting {
			feature.Endpoints = make(map[int]*models.Endpoint)
			role.Entries[feature.ID] = &feature
		}

		role.Entries[feature.ID].Endpoints[endpoint.ID] = &endpoint
	}

	// rows.Scan after db.Query doesn't return sql.ErrNoRows
	if role.ID == 0 {
		return &models.Role{}, ErrNoRows
	}

	err = rows.Err()
	if err != nil {
		return &models.Role{}, ScanError{scanErrorMessage, err}
	}

	return &role, nil
}

func GetRoleTemplate(repository *RolesRepository) (*models.Role, error) {
	query := `SELECT features.id, features.name, endpoints.id, endpoints.path, endpoints.method
			  FROM features LEFT JOIN "featuresToEndpoints"
			  ON features.id = "featuresToEndpoints"."featureId"
			  LEFT JOIN endpoints
			  ON "featuresToEndpoints"."endpointId" = endpoints.id`

	rows, err := repository.DB.Query(query)
	if err != nil {
		return &models.Role{}, QueryError{queryErrorMessage, err}
	}

	var roleTemplate models.Role
	var features = make(map[int]*models.FeatureEntry)

	// Get all features and connect them to endpoints
	for rows.Next() {
		var feature models.FeatureEntry
		var endpoint models.Endpoint
		isFeatureExisting := false

		err := rows.Scan(&feature.ID, &feature.Name, &endpoint.ID, &endpoint.Method, &endpoint.Path)
		if err != nil {
			return &models.Role{}, ScanError{scanErrorMessage, err}
		}

		isFeatureExisting = false

		_, isFeatureExisting = features[feature.ID]
		if !isFeatureExisting {
			feature.Endpoints = make(map[int]*models.Endpoint)
			features[feature.ID] = &feature
		}

		features[feature.ID].Endpoints[endpoint.ID] = &endpoint
	}

	err = rows.Err()
	if err != nil {
		return &models.Role{}, ScanError{scanErrorMessage, err}
	}

	roleTemplate.Entries = features

	return &roleTemplate, nil
}

func CreateRole(repository *RolesRepository, role *models.Role) error {
	query := `INSERT INTO roles(name) VALUES($1) RETURNING(id)`

	var roleId int

	err := repository.DB.QueryRow(query, role.Name).Scan(&roleId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRows
		} else {
			err = QueryError{queryErrorMessage, err}
		}
		return err
	}

	// Establish connection between the role and it's features
	query = `INSERT INTO "rolesToFeatures"("roleId", "featureId") VALUES($1, $2)`

	for _, feature := range role.Entries {
		_, err = repository.DB.Exec(query, roleId, feature.ID)
		if err != nil {
			return QueryError{queryErrorMessage, err}
		}
	}

	return nil
}

func UpdateRole(repository *RolesRepository, role *models.Role, id int) error {
	query := `SELECT FROM roles WHERE id = $1`

	err := repository.DB.QueryRow(query, id).Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRows
		} else {
			err = QueryError{queryErrorMessage, err}
		}
		return err
	}

	// Update role's id and name
	query = `UPDATE roles SET name = $1 WHERE id = $2`

	_, err = repository.DB.Exec(query, role.Name, id)
	if err != nil {
		return QueryError{queryErrorMessage, err}
	}

	// Delete all connections with the role
	query = `DELETE FROM "rolesToFeatures" WHERE "roleId" = $1`

	_, err = repository.DB.Exec(query, id)
	if err != nil {
		return QueryError{queryErrorMessage, err}
	}

	// Establish new connection between the role and it's features
	query = `INSERT INTO "rolesToFeatures"("roleId", "featureId") VALUES ($1, $2)`

	for _, feature := range role.Entries {
		_, err = repository.DB.Exec(query, id, feature.ID)
		if err != nil {
			return QueryError{queryErrorMessage, err}
		}
	}

	return nil
}

func DeleteRole(repository *RolesRepository, id int) error {
	query := `SELECT FROM roles WHERE id = $1`

	err := repository.DB.QueryRow(query, id).Scan()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNoRows
		} else {
			err = QueryError{queryErrorMessage, err}
		}
		return err
	}

	query = `DELETE FROM roles WHERE id = $1`

	_, err = repository.DB.Exec(query, id)
	if err != nil {
		return QueryError{queryErrorMessage, err}
	}

	return nil
}

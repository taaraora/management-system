BEGIN;

CREATE TABLE IF NOT EXISTS endpoints (
    id SERIAL PRIMARY KEY,
    path character varying(255) NOT NULL,
    method character varying(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS features (
    id SERIAL PRIMARY KEY,
    name character varying(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    id integer PRIMARY KEY,
    name character varying(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS "featuresToEndpoints" (
    "featureId" integer NOT NULL REFERENCES features(id),
    "endpointId" integer NOT NULL REFERENCES endpoints(id)
);

CREATE TABLE IF NOT EXISTS "rolesToFeatures" (
    "roleId" integer NOT NULL REFERENCES roles(id) ON UPDATE CASCADE ON DELETE CASCADE,
    "featureId" integer NOT NULL REFERENCES features(id) ON UPDATE CASCADE ON DELETE CASCADE
);

COMMIT;
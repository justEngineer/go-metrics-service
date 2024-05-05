CREATE TABLE gauge_metrics (
		id      VARCHAR (255) PRIMARY KEY,
		value   DOUBLE PRECISION NOT NULL
	);
	CREATE TABLE counter_metrics (
		id      VARCHAR (255) PRIMARY KEY,
		value   BIGSERIAL NOT NULL
	);
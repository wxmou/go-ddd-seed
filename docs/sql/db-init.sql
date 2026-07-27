-- public.audit_logs 定义

-- Drop table

-- DROP TABLE audit_logs;

CREATE TABLE audit_logs (
                            id varchar(36) NOT NULL,
                            operator_id varchar(36) NOT NULL,
                            operator_name varchar(100) DEFAULT ''::character varying NOT NULL,
                            "action" varchar(32) NOT NULL,
                            target_type varchar(64) NOT NULL,
                            target_id varchar(36) DEFAULT ''::character varying NOT NULL,
                            request_body text DEFAULT ''::text NOT NULL,
                            response_body text DEFAULT ''::text NOT NULL,
                            client_ip varchar(45) DEFAULT ''::character varying NOT NULL,
                            user_agent varchar(512) DEFAULT ''::character varying NOT NULL,
                            trace_id varchar(36) DEFAULT ''::character varying NOT NULL,
                            created_at timestamptz NULL,
                            CONSTRAINT audit_logs_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_audit_action ON public.audit_logs USING btree (action);
CREATE INDEX idx_audit_created ON public.audit_logs USING btree (created_at);
CREATE INDEX idx_audit_logs_trace_id ON public.audit_logs USING btree (trace_id);
CREATE INDEX idx_audit_operator ON public.audit_logs USING btree (operator_id);
CREATE INDEX idx_audit_target ON public.audit_logs USING btree (target_type);


-- public.dict_entries 定义

-- Drop table

-- DROP TABLE dict_entries;

CREATE TABLE dict_entries (
                              id varchar(36) NOT NULL,
                              type_id varchar(36) NOT NULL,
                              "label" varchar(200) NOT NULL,
                              value varchar(200) NOT NULL,
                              sort_order int8 DEFAULT 0 NULL,
                              status int8 DEFAULT 1 NULL,
                              created_at timestamptz NULL,
                              updated_at timestamptz NULL,
                              CONSTRAINT dict_entries_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_dict_entries_type_id ON public.dict_entries USING btree (type_id);


-- public.dict_types 定义

-- Drop table

-- DROP TABLE dict_types;

CREATE TABLE dict_types (
                            id varchar(36) NOT NULL,
                            code varchar(100) NOT NULL,
                            "name" varchar(200) NOT NULL,
                            description varchar(500) NULL,
                            status int8 DEFAULT 1 NULL,
                            created_at timestamptz NULL,
                            updated_at timestamptz NULL,
                            CONSTRAINT dict_types_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_dict_types_code ON public.dict_types USING btree (code);


-- public.field_configs 定义

-- Drop table

-- DROP TABLE field_configs;

CREATE TABLE field_configs (
                               id varchar(36) NOT NULL,
                               entity_type varchar(100) NOT NULL,
                               field_name varchar(100) NOT NULL,
                               field_label varchar(200) NOT NULL,
                               field_type varchar(50) NOT NULL,
                               required bool DEFAULT false NULL,
                               sort_order int8 DEFAULT 0 NULL,
                               "options" text NULL,
                               status int8 DEFAULT 1 NULL,
                               created_at timestamptz NULL,
                               updated_at timestamptz NULL,
                               CONSTRAINT field_configs_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_entity_field ON public.field_configs USING btree (entity_type, field_name);


-- public.file_records 定义

-- Drop table

-- DROP TABLE file_records;

CREATE TABLE file_records (
                              id varchar(36) NOT NULL,
                              file_name varchar(255) NOT NULL,
                              storage_path varchar(1024) NOT NULL,
                              "size" int8 NOT NULL,
                              mime_type varchar(127) NULL,
                              storage_channel varchar(32) NULL,
                              md5_hash varchar(64) NULL,
                              attach_type varchar(64) NULL,
                              attach_id varchar(36) NULL,
                              uploader_id varchar(36) NULL,
                              thumbnail_path varchar(1024) NULL,
                              status int8 DEFAULT 1 NULL,
                              created_at timestamptz NULL,
                              updated_at timestamptz NULL,
                              CONSTRAINT file_records_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_file_attach ON public.file_records USING btree (attach_type, attach_id);
CREATE INDEX idx_file_created ON public.file_records USING btree (created_at);
CREATE INDEX idx_file_uploader ON public.file_records USING btree (uploader_id);


-- public.kv_configs 定义

-- Drop table

-- DROP TABLE kv_configs;

CREATE TABLE kv_configs (
                            id varchar(36) NOT NULL,
                            "key" varchar(255) NOT NULL,
                            value text NOT NULL,
                            description varchar(500) NULL,
                            status int8 DEFAULT 1 NULL,
                            created_at timestamptz NULL,
                            updated_at timestamptz NULL,
                            CONSTRAINT kv_configs_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_kv_configs_key ON public.kv_configs USING btree (key);


-- public.permissions 定义

-- Drop table

-- DROP TABLE permissions;

CREATE TABLE permissions (
                             id varchar(36) NOT NULL,
                             "name" varchar(200) NOT NULL,
                             code varchar(200) NOT NULL,
                             description varchar(500) NULL,
                             created_at timestamptz NULL,
                             updated_at timestamptz NULL,
                             CONSTRAINT permissions_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_permissions_code ON public.permissions USING btree (code);


-- public.role_permissions 定义

-- Drop table

-- DROP TABLE role_permissions;

CREATE TABLE role_permissions (
                                  id varchar(36) NOT NULL,
                                  role_id varchar(36) NOT NULL,
                                  permission_id varchar(36) NOT NULL,
                                  created_at timestamptz NULL,
                                  CONSTRAINT role_permissions_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX uk_role_permission ON public.role_permissions USING btree (role_id, permission_id);


-- public.roles 定义

-- Drop table

-- DROP TABLE roles;

CREATE TABLE roles (
                       id varchar(36) NOT NULL,
                       "name" varchar(100) NOT NULL,
                       code varchar(100) NOT NULL,
                       description varchar(500) NULL,
                       status int8 DEFAULT 1 NULL,
                       created_at timestamptz NULL,
                       updated_at timestamptz NULL,
                       CONSTRAINT roles_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_roles_code ON public.roles USING btree (code);


-- public.user_roles 定义

-- Drop table

-- DROP TABLE user_roles;

CREATE TABLE user_roles (
                            id varchar(36) NOT NULL,
                            user_id varchar(36) NOT NULL,
                            role_id varchar(36) NOT NULL,
                            created_at timestamptz NULL,
                            CONSTRAINT user_roles_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX uk_user_role ON public.user_roles USING btree (user_id, role_id);


-- public.users 定义

-- Drop table

-- DROP TABLE users;

CREATE TABLE users (
                       id varchar(36) NOT NULL,
                       username varchar(100) NOT NULL,
                       password_hash varchar(255) NOT NULL,
                       real_name varchar(100) NULL,
                       email varchar(200) NULL,
                       phone varchar(50) NULL,
                       status int8 DEFAULT 1 NULL,
                       last_login_at timestamptz NULL,
                       created_at timestamptz NULL,
                       updated_at timestamptz NULL,
                       CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX idx_users_username ON public.users USING btree (username);
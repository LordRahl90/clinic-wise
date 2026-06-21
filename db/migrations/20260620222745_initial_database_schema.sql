-- +goose Up
create table audit_trails
(
    id             varbinary(50) not null primary key,
    actor_id       varbinary(50) null,
    actor_name     longtext      null,
    actor_role     longtext      null,
    action         longtext      null,
    entity_type    longtext      null,
    entity_id      varchar(50)   null,
    appointment_id varchar(50)   null,
    message        text          null,
    changes        json          null,
    created_at     datetime(3)   null,
    updated_at     datetime(3)   null,
    deleted_at     datetime(3)   null
);

create index idx_audit_trails_deleted_at
    on audit_trails (deleted_at);

create table diagnoses
(
    id             varbinary(50)        not null primary key,
    hospital_id    longblob             null,
    doctor_id      longblob             null,
    patient_id     longblob             null,
    appointment_id longblob             null,
    diagnosis      longtext             null,
    details        longtext             null,
    dismissed      tinyint(1) default 0 null,
    created_at     datetime(3)          null,
    updated_at     datetime(3)          null,
    deleted_at     datetime(3)          null
);

create index idx_diagnoses_deleted_at
    on diagnoses (deleted_at);

create table hospitals
(
    id         varbinary(50) not null
        primary key,
    name       longtext      null,
    created_at datetime(3)   null,
    updated_at datetime(3)   null,
    deleted_at datetime(3)   null
);

create index idx_hospitals_deleted_at
    on hospitals (deleted_at);

create table notes
(
    id             varbinary(50) not null
        primary key,
    hospital_id    longblob      null,
    appointment_id longblob      null,
    doctor_id      longblob      null,
    patient_id     longblob      null,
    content        longtext      null,
    created_at     datetime(3)   null,
    updated_at     datetime(3)   null,
    deleted_at     datetime(3)   null
);

create index idx_notes_deleted_at
    on notes (deleted_at);

create table prescriptions
(
    id              varbinary(50) not null primary key,
    hospital_id     longblob      null,
    doctor_id       longblob      null,
    patient_id      longblob      null,
    appointment_id  longblob      null,
    expiration_date datetime(3)   null,
    details         longtext      null,
    status          longtext      null,
    created_at      datetime(3)   null,
    updated_at      datetime(3)   null,
    deleted_at      datetime(3)   null
);

create index idx_prescriptions_deleted_at
    on prescriptions (deleted_at);

create table timeslots
(
    id          varbinary(50) not null
        primary key,
    hospital_id longblob      null,
    user_id     longblob      null,
    date        longtext      null,
    start_time  longtext      null,
    end_time    longtext      null,
    created_at  datetime(3)   null,
    updated_at  datetime(3)   null,
    deleted_at  datetime(3)   null
);

create index idx_timeslots_deleted_at
    on timeslots (deleted_at);

create table users
(
    id          varbinary(50)        not null
        primary key,
    hospital_id varbinary(50)        null,
    first_name  longtext             null,
    last_name   longtext             null,
    email       varchar(255)         null,
    password    varchar(255)         null,
    role        longtext             null,
    accepted    tinyint(1) default 0 null,
    created_at  datetime(3)          null,
    updated_at  datetime(3)          null,
    deleted_at  datetime(3)          null,
    constraint idx_users_email
        unique (email)
);

create table appointments
(
    id          varbinary(50) not null
        primary key,
    hospital_id longblob      null,
    doctor_id   varbinary(50) null,
    patient_id  varbinary(50) null,
    timeslot_id longblob      null,
    description longtext      null,
    status      longtext      null,
    created_at  datetime(3)   null,
    updated_at  datetime(3)   null,
    deleted_at  datetime(3)   null,
    constraint fk_appointments_doctor
        foreign key (doctor_id) references users (id),
    constraint fk_appointments_patient
        foreign key (patient_id) references users (id)
);

create index idx_appointments_deleted_at
    on appointments (deleted_at);

create index idx_users_deleted_at
    on users (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS audit_trails;
DROP TABLE IF EXISTS diagnoses;
DROP TABLE IF EXISTS hospitals;
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS prescriptions;
DROP TABLE IF EXISTS timeslots;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS appointments;


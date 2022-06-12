create table Users (
    UserId bigint generated always as identity not null,
    EmailAddress varchar(255) not null,
    Activated boolean not null,
    RecordedAt timestamp not null, 
    UpdatedAt timestamp not null,
    primary key (UserId)
);

create index IDX_Users_EmailAddress on Users (EmailAddress);

create table PasswordHashes (
    UserId bigint not null,
    PasswordHash varchar(100) not null,
    RecordedAt timestamp not null,
    UpdatedAt timestamp not null,
    primary key (PasswordHash)
);
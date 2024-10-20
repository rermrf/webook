create database webook;

use webook;

CREATE TABLE `users` (
                         `id` bigint NOT NULL AUTO_INCREMENT,
                         `email` varchar(191) DEFAULT NULL,
                         `phone` varchar(191) DEFAULT NULL,
                         `password` longtext,
                         `ctime` bigint DEFAULT NULL,
                         `utime` bigint DEFAULT NULL,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uni_users_email` (`email`),
                         UNIQUE KEY `uni_users_phone` (`phone`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

create table `articles`
(
    id        bigint auto_increment
        primary key,
    title     varchar(1024)    null,
    content   blob             null,
    author_id bigint           null,
    status    tinyint unsigned null,
    ctime     bigint           null,
    utime     bigint           null
);

create index idx_articles_author_id
    on webook.articles (author_id);

create table interactives
(
    id          bigint auto_increment
        primary key,
    biz_id      bigint       null,
    biz         varchar(128) null,
    read_cnt    bigint       null,
    like_cnt    bigint       null,
    collect_cnt bigint       null,
    ctime       bigint       null,
    utime       bigint       null,
    constraint biz_id_type
        unique (biz_id, biz)
);

create table webook.published_articles
(
    id        bigint auto_increment
        primary key,
    title     varchar(1024)    null,
    content   blob             null,
    author_id bigint           null,
    status    tinyint unsigned null,
    ctime     bigint           null,
    utime     bigint           null
);

create index idx_published_articles_author_id
    on webook.published_articles (author_id);


create table webook.user_collection_bizs
(
    id     bigint auto_increment
        primary key,
    cid    bigint       null,
    biz_id bigint       null,
    biz    varchar(128) null,
    uid    bigint       null,
    ctime  bigint       null,
    utime  bigint       null,
    constraint biz_type_id_uid
        unique (biz_id, biz, uid)
);

create index idx_user_collection_bizs_cid
    on webook.user_collection_bizs (cid);


create table webook.user_like_bizs
(
    id     bigint auto_increment
        primary key,
    biz    varchar(128)     null,
    biz_id bigint           null,
    uid    bigint           null,
    ctime  bigint           null,
    utime  bigint           null,
    status tinyint unsigned null,
    constraint uid_biz_id_type
        unique (biz, biz_id, uid)
);


create table webook.users
(
    id              bigint auto_i
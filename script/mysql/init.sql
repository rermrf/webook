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


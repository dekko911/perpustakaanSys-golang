CREATE TABLE
    IF NOT EXISTS `users` (
        `id` CHAR(36) NOT NULL,
        `name` VARCHAR(255) NOT NULL,
        `email` VARCHAR(255) NOT NULL,
        `password` VARCHAR(100) NOT NULL,
        `avatar` VARCHAR(100) NULL,
        `token_version` INT NOT NULL DEFAULT 1,
        `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`id`),
        UNIQUE KEY (`email`)
    );
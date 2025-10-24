CREATE TABLE
    IF NOT EXISTS `role_user` (
        `user_id` CHAR(36) NOT NULL,
        `role_id` CHAR(36) NOT NULL,
        `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        PRIMARY KEY (`user_id`, `role_id`),
        CONSTRAINT `fk_role_user_user_id` FOREIGN KEY (`user_id`) REFERENCES users (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
        CONSTRAINT `fk_role_user_role_id` FOREIGN KEY (`role_id`) REFERENCES roles (`id`) ON DELETE CASCADE ON UPDATE CASCADE
    );
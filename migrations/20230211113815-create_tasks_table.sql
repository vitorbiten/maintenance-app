-- +migrate Up
CREATE TABLE IF NOT EXISTS `tasks` (
  `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT,
  `summary` text NOT NULL,
  `author_id` bigint(10) unsigned NOT NULL,
  `date` datetime DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `tasks_author_id_users_id_foreign` (`author_id`),
  CONSTRAINT `tasks_author_id_users_id_foreign` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- +migrate Down
DROP TABLE `tasks`;
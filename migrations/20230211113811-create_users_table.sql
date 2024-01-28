-- +migrate Up
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT,
  `nickname` varchar(255) NOT NULL,
  `email` varchar(100) NOT NULL,
  `user_type` enum('manager','technician') DEFAULT 'technician',
  `password` varchar(100) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `nickname` (`nickname`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO maintenance_api.users
(id, nickname, email, user_type, password, created_at, updated_at)
VALUES(0, 'Martin Luther', 'luther@gmail.com', 'manager', 'password', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- +migrate Down
DROP TABLE `users`;
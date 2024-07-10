DROP TABLE IF EXISTS `unique_request`;
CREATE TABLE `unique_request`
(
    `id`          bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `request_id`  char(32)            NOT NULL COMMENT '对成功幂等',
    `task_id`     char(32)            NOT NULL,
    `create_time` timestamp           NOT NULL DEFAULT current_timestamp(),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_request_id` (`request_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb3
  COLLATE = utf8mb3_general_ci COMMENT ='防重表，必须，创建更新操作对成功幂等';

DROP TABLE IF EXISTS `task`;
CREATE TABLE `task`
(
    `id`          char(32)     NOT NULL,
    `request_id`  char(32)     NOT NULL COMMENT '初始请求ID',
    `type`        varchar(128) NOT NULL COMMENT '业务类型',
    `state`       varchar(128) NOT NULL COMMENT '任务状态',
    `version`     int(11)      NOT NULL DEFAULT 1,
    `create_time` timestamp    NOT NULL DEFAULT current_timestamp(),
    `update_time` timestamp    NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_request_id` (`request_id`),
    KEY `idx_state` (`state`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb3
  COLLATE = utf8mb3_general_ci COMMENT ='任务主表，必须，维护状态驱动执行';


DROP TABLE IF EXISTS `data`;
CREATE TABLE `data`
(
    `id`               bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `task_id`          char(32)            NOT NULL,
    -- 以下是业务字段，例如
    `symbol`           varchar(20)         NOT NULL DEFAULT '',
    `quantity`         decimal(50, 15)     NOT NULL DEFAULT 0.000000000000000,
    `amount`           decimal(50, 15)     NOT NULL DEFAULT 0.000000000000000,
    `operator`         varchar(20)         NOT NULL DEFAULT '',
    `transaction_time` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '业务时间',
    `comment`          varchar(128)        NOT NULL DEFAULT '' COMMENT '备注说明',
    PRIMARY KEY (`id`),
    KEY `idx_task_id` (`task_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb3
  COLLATE = utf8mb3_general_ci COMMENT ='业务字段表，必须，根据具体业务设计字段';

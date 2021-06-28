create table fetch_record
(
    id         int(10) auto_increment	primary key COMMENT '自增主键',
    update_at  varchar(36)  not null COMMENT '发布日期',
    update_time varchar(36)  null COMMENT '抓取时间',
    down_url   varchar(100) not null COMMENT 'oss链接'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='国家统计局数据存储oss';
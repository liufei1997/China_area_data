create table special_region
(
    id          int(10) auto_increment comment '自增主键'
        primary key,
    zoning_code varchar(36)  not null comment '统计区划代码',
    city_code   varchar(36) null comment '市级code',
    region_name varchar(100) not null comment '区级名称',
    replace_id  int(10) not null comment '区级code末尾2位',
    create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='东莞市,中山市和儋州等市数据';

ALTER TABLE `special_region`
    ADD UNIQUE (`zoning_code`);
ALTER TABLE `special_region`
    ADD INDEX index_name (`city_code`);
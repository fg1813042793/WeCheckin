ALTER TABLE `workflow_definitions`
  ADD COLUMN `definition_logo_url` varchar(500) NOT NULL DEFAULT '' COMMENT '流程Logo地址' AFTER `definition_category`;

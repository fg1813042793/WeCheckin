ALTER TABLE `workflow_process_instances`
  ADD COLUMN `form_data_json` mediumtext COMMENT '流程表单数据JSON' AFTER `instance_status`;

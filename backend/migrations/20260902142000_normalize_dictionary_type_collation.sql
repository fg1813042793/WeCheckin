-- Align dictionary type metadata with the historical sys_dicts collation.
-- This repairs databases that already ran 20260902113000 before its collation was explicit.

ALTER TABLE `sys_dict_types`
  CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

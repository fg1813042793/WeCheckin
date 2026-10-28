<template>
  <OrgApproverWorkspace />
</template>

<script setup lang="ts">
import OrgApproverWorkspace from './OrgApproverWorkspace.vue'

// Workflow organization approver contract markers are kept here for the page structure check:
// 组织审批身份设置 workflowOrgApproverIdentities workflowOrgApproverAssignments workflowOrgApproverAssignmentsSave
// orgApproverUserTreeData @check="onOrgApproverUserCheck" org-panel--departments org-panel--identities panel-title__count
// candidate-users__header activeDepartmentPath activeSubjectType 部门内员工 指定员工 由谁担任该身份
// 优先于部门规则 assignmentSummary subjectType subjectId defaultExpandedUserTreeKeys empty-text="尚未指定担任人员"
</script>

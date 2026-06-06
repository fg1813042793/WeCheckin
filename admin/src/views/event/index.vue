<template>
  <div>
    <el-card>
      <div style="display:flex;gap:10px;margin-bottom:12px">
        <el-input v-model="keyword" placeholder="搜索标题" clearable style="width:200px" @keyup.enter="search" />
        <el-select v-model="typeFilter" placeholder="全部类型" clearable style="width:120px" @change="search">
          <el-option label="活动" :value="1" />
          <el-option label="赛事" :value="2" />
        </el-select>
        <el-button type="primary" @click="search">搜索</el-button>
      </div>
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button v-if="hasPerm('event:add')" type="success" @click="showAdd">+ 新增</el-button>
          <el-button v-if="hasPerm('event:del')" type="danger" :disabled="selected.length === 0" @click="delSelected">批量删除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-button circle icon="Refresh" title="刷新" @click="load" />
          <el-button circle icon="Upload" title="导入" @click="ElMessage.info('导入功能开发中')" />
          <el-button circle icon="Download" title="导出" @click="exportData" />
          <SortPopover :columns="sortColumns" v-model="sortRules" @change="onSortChange" />
        </div>
      </div>
      <el-table :data="list" v-loading="loading" stripe style="width:100%" @selection-change="selected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="title" label="标题" min-width="160" />
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 1 ? 'warning' : 'danger'" size="small">{{ row.type === 1 ? '活动' : '赛事' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报名时间" min-width="200">
          <template #default="{ row }">{{ row.regStartStr || '-' }} ~ {{ row.regEndStr || '-' }}</template>
        </el-table-column>
        <el-table-column label="活动时间" min-width="200">
          <template #default="{ row }">{{ row.eventStartStr || '-' }} ~ {{ row.eventEndStr || '-' }}</template>
        </el-table-column>
        <el-table-column label="标记" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.vouch" type="success" size="small">推荐</el-tag>
            <el-tag v-if="row.isTop" type="danger" size="small" style="margin-left:4px">置顶</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="userCnt" label="参与人数" width="80" />
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button v-if="hasPerm('event:edit')" size="small" type="primary" @click="showEdit(row)">编辑</el-button>
              <el-button v-if="hasPerm('event:list')" size="small" @click="showParticipants(row)">参与者</el-button>
              <el-button v-if="hasPerm('event:list')" size="small" @click="showDynamics(row)">动态</el-button>
              <el-dropdown v-if="hasPerm('event:edit')" trigger="click" @command="(cmd:string)=>handleMore(cmd,row)">
                <el-button size="small">更多<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="enable" :disabled="row.status===1">启用</el-dropdown-item>
                      <el-dropdown-item command="disable" :disabled="row.status===0">停用</el-dropdown-item>
                      <el-dropdown-item :command="row.vouch ? 'unvouch' : 'vouch'">{{ row.vouch ? '取消推荐' : '推荐首页' }}</el-dropdown-item>
                      <el-dropdown-item :command="row.isTop ? 'untop' : 'top'">{{ row.isTop ? '取消置顶' : '置顶' }}</el-dropdown-item>
                      <el-dropdown-item command="scores" v-if="row.type===2 && hasPerm('event:list')">成绩</el-dropdown-item>
                      <el-dropdown-item v-if="hasPerm('event:del')" command="del" divided>删除</el-dropdown-item>
                    </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div style="text-align:center;margin-top:16px">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :page-sizes="[10,20,50,100]" :total="total" layout="total,sizes,prev,pager,next" @current-change="load" @size-change="(val:number) => { pageSize = val; page = 1; load() }" />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formDialog.visible" :title="formDialog.title" width="800px" :close-on-click-modal="false" destroy-on-close>
      <el-form ref="formRef" :model="form" label-width="120px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="必填" />
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="form.type">
            <el-radio :value="1">活动</el-radio>
            <el-radio :value="2">赛事</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="封面">
          <el-upload action="/upload" :show-file-list="false" :on-success="handleCoverSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="cover-upload">
              <el-image v-if="form.cover" :src="form.cover" class="cover-preview" />
              <div v-else class="cover-placeholder">+</div>
              <div v-if="form.cover" class="cover-overlay" @click.stop>
                <el-button size="small" type="danger" :icon="Delete" circle @click.stop="form.cover=''" />
              </div>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.desc" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="form.type === 2 ? '赛事规则' : '活动规则'">
          <el-input v-model="form.rules" type="textarea" :rows="4" placeholder="填写活动/赛事的规则说明" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.cateName" style="width:200px">
            <el-option v-for="c in categories" :key="c.value" :label="c.label" :value="c.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="报名开始">
          <el-date-picker v-model="form.regStart" type="datetime" value-format="x" style="width:200px" />
        </el-form-item>
        <el-form-item label="报名结束">
          <el-date-picker v-model="form.regEnd" type="datetime" value-format="x" style="width:200px" />
        </el-form-item>
        <el-form-item label="活动开始">
          <el-date-picker v-model="form.eventStart" type="datetime" value-format="x" style="width:200px" />
        </el-form-item>
        <el-form-item label="活动结束">
          <el-date-picker v-model="form.eventEnd" type="datetime" value-format="x" style="width:200px" />
        </el-form-item>
        <el-form-item label="发布部门">
          <el-popover trigger="click" placement="bottom" :width="260" popper-style="padding:0">
            <template #reference>
              <div class="multi-select-input">
                <el-tag v-for="id in (form.publishDeptIds ? form.publishDeptIds.split(',').map(Number) : [])" :key="id" size="small" closable @close.stop="form.publishDeptIds=form.publishDeptIds.split(',').map(Number).filter((k:any)=>k!==id).join(',')">{{ getDeptPath(id) }}</el-tag>
                <span class="ms-placeholder" v-if="!form.publishDeptIds">选择发布部门</span>
                <el-icon class="ms-arrow"><ArrowDown /></el-icon>
              </div>
            </template>
            <el-tree ref="deptTreeRef" :data="deptTree" :props="{ label: 'name', children: 'children' }" node-key="id" show-checkbox check-strictly :default-checked-keys="form.publishDeptIds ? form.publishDeptIds.split(',').map(Number) : []" @check="(data:any,{checkedKeys}:any)=>{form.publishDeptIds=checkedKeys.filter((k:any)=>k!==0).join(',')}" />
          </el-popover>
        </el-form-item>
        <el-form-item label="主办人">
          <div class="multi-select-input" @click="showUserPicker('organizer')">
            <el-tag v-for="(u, i) in form.organizers" :key="i" size="small" closable @close.stop="form.organizers.splice(i,1)">{{ u.name }}</el-tag>
            <span class="ms-placeholder" v-if="form.organizers.length === 0">请选择主办人</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
        <el-form-item label="主办人助理" v-if="form.type">
          <div class="multi-select-input" @click="showUserPicker('assistant')">
            <el-tag v-for="(u, i) in form.assistants" :key="i" size="small" closable @close.stop="form.assistants.splice(i,1)">{{ u.name }}</el-tag>
            <span class="ms-placeholder" v-if="form.assistants.length === 0">请选择助理</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
        <el-form-item label="裁判" v-if="form.type === 2">
          <div class="multi-select-input" @click="showUserPicker('referee')">
            <el-tag v-for="(u, i) in form.referees" :key="i" size="small" closable @close.stop="form.referees.splice(i,1)">{{ u.name }}</el-tag>
            <span class="ms-placeholder" v-if="form.referees.length === 0">请选择裁判</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
        <el-form-item label="报名表单">
          <div class="multi-select-input" @click="showFormEditor">
            <el-tag v-for="(f, i) in form.fields" :key="i" size="small" closable @close.stop="form.fields.splice(i,1)">{{ f.label || '字段' + (i+1) }}</el-tag>
            <span class="ms-placeholder" v-if="form.fields.length === 0">点击配置报名表单字段</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
        <el-form-item label="评分项" v-if="form.type === 2">
          <div class="multi-select-input" @click="showScoreFieldEditor">
            <el-tag v-for="(sf, i) in form.scoreFields" :key="i" size="small" closable @close.stop="form.scoreFields.splice(i,1)">{{ sf.name || '评分项' + (i+1) }}</el-tag>
            <span class="ms-placeholder" v-if="form.scoreFields.length === 0">点击配置评分项</span>
            <el-icon class="ms-arrow"><ArrowDown /></el-icon>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEvent">{{ formDialog.isCreate ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <!-- 人员选择弹窗 -->
    <el-dialog v-model="userPicker.visible" :title="'选择' + userPickerTitle" width="600px">
      <div style="display:flex;gap:12px">
        <div class="dept-tree-wrap">
          <div class="dept-tree-title">部门</div>
          <el-tree :data="deptTree" :props="{ label: 'name', children: 'children' }" node-key="id" highlight-current @node-click="handleDeptClick" />
        </div>
        <div class="dept-users-wrap">
          <div class="dept-tree-title">用户</div>
          <el-table :data="userPicker.users" stripe max-height="360" @selection-change="(sel:any)=>userPicker.selected=sel">
            <el-table-column type="selection" width="40" />
            <el-table-column label="用户" min-width="160">
              <template #default="{ row }">{{ row.name || row.openid }}</template>
            </el-table-column>
          </el-table>
        </div>
      </div>
      <template #footer>
        <el-button @click="confirmUserPicker">确认选择 ({{ userPicker.selected.length }} 人)</el-button>
      </template>
    </el-dialog>

    <!-- 表单字段配置弹窗 -->
    <el-dialog v-model="formEditor.visible" title="配置报名表单字段" width="600px">
      <div v-for="(f, i) in formEditor.fields" :key="i" style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
        <span style="color:#999">{{ i + 1 }}.</span>
        <el-input v-model="f.label" placeholder="字段名称" style="width:140px" />
        <el-select v-model="f.type" style="width:120px">
          <el-option label="文本" value="text" />
          <el-option label="数字" value="number" />
          <el-option label="多行文本" value="textarea" />
          <el-option label="选择" value="select" />
          <el-option label="拍照上传" value="image" />
          <el-option label="位置签到" value="location" />
        </el-select>
        <el-input v-if="f.type==='select'" v-model="f.options" placeholder="选项(逗号分隔)" style="width:160px" />
        <el-checkbox v-model="f.required">必填</el-checkbox>
        <el-button type="danger" :icon="Delete" circle size="small" @click="formEditor.fields.splice(i,1)" />
      </div>
      <el-button @click="formEditor.fields.push({label:'',type:'text',options:'',required:false})">+ 添加字段</el-button>
      <template #footer>
        <el-button @click="confirmFormEditor">完成</el-button>
      </template>
    </el-dialog>

    <!-- 评分项配置弹窗 -->
    <el-dialog v-model="scoreFieldEditor.visible" title="配置评分项" width="600px">
      <div v-for="(sf, i) in scoreFieldEditor.fields" :key="i" style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
        <span style="color:#999">{{ i + 1 }}.</span>
        <el-input v-model="sf.name" placeholder="评分项名称" style="width:140px" />
        <el-select v-model="sf.type" style="width:100px" @change="sf.options=''">
          <el-option label="数字" value="number" />
          <el-option label="文本" value="text" />
          <el-option label="选择" value="select" />
        </el-select>
        <el-input v-if="sf.type==='select'" v-model="sf.options" placeholder="选项(逗号分隔)" style="width:160px" />
        <el-button type="danger" :icon="Delete" circle size="small" @click="scoreFieldEditor.fields.splice(i,1)" />
      </div>
      <el-button @click="scoreFieldEditor.fields.push({name:'',type:'number',options:''})">+ 添加评分项</el-button>
      <template #footer>
        <el-button @click="confirmScoreFieldEditor">完成</el-button>
      </template>
    </el-dialog>

    <!-- 参与者列表 -->
    <el-dialog v-model="partDialog.visible" :title="partDialog.title" width="1000px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button type="danger" :disabled="partSelected.length === 0" @click="delSelectedParticipants">批量删除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-button circle icon="Download" title="导出 CSV" @click="exportParticipants" />
        </div>
      </div>
      <el-table :data="partList" v-loading="partLoading" stripe @selection-change="partSelected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column label="用户" min-width="140">
          <template #default="{ row }">{{ row.userName || row.miniOpenId }}</template>
        </el-table-column>
        <el-table-column label="部门" min-width="120">
          <template #default="{ row }">{{ row.deptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="顶级部门" min-width="120">
          <template #default="{ row }">{{ row.topDeptName || '-' }}</template>
        </el-table-column>
        <el-table-column label="报名时间" width="180">
          <template #default="{ row }">{{ fmtTime(row._createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button size="small" type="danger" link @click="delParticipant(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 动态列表 -->
    <el-dialog v-model="dynDialog.visible" :title="dynDialog.title" width="1200px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div class="toolbar-icons">
          <el-button size="small" type="primary" @click="addDynamic">添加动态</el-button>
          <el-button size="small" type="danger" :disabled="dynSelected.length === 0" @click="delSelectedDynamics">批量删除</el-button>
        </div>
        <div class="toolbar-icons">
          <el-button circle icon="Upload" title="导入" size="small" @click="ElMessage.info('导入功能开发中')" />
          <el-button circle icon="Download" title="导出" size="small" @click="ElMessage.info('导出功能开发中')" />
          <el-button circle icon="Sort" title="排序" size="small" @click="ElMessage.info('排序功能开发中')" />
        </div>
      </div>
      <el-table :data="dynList" stripe @selection-change="dynSelected = $event">
        <el-table-column type="selection" width="45" />
        <el-table-column label="发布者" width="100">
          <template #default="{ row }">{{ row.userName || row.userId }}</template>
        </el-table-column>
        <el-table-column label="标题" width="100">
          <template #default="{ row }">{{ row.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="内容" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column label="图片" width="160">
          <template #default="{ row }">
            <div v-if="row.imageList && row.imageList.length > 0" class="scroll-x" @wheel="onWheel">
              <el-image v-for="(img, j) in row.imageList" :key="j" :src="img" style="width:50px;height:50px;border-radius:4px;flex-shrink:0" fit="cover" :preview-src-list="row.imageList" preview-teleported />
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="视频" width="160">
          <template #default="{ row }">
            <div v-if="row.videoList && row.videoList.length > 0" class="scroll-x" @wheel="onWheel">
              <div v-for="(vurl, j) in row.videoList" :key="j" :title="vurl" :style="{ width:'50px', height:'50px', borderRadius:'4px', cursor:'pointer', flexShrink:0, background:`#eee url(${getVideoThumb(vurl)}) center/cover` }" @click="previewVideo(vurl)">
                <div style="width:100%;height:100%;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,.3);border-radius:4px;color:#fff;font-size:18px">▶</div>
              </div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="140">
          <template #default="{ row }">{{ fmtTime(row._createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="110" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="editDynamic(row)">编辑</el-button>
            <el-button size="small" type="danger" link @click="delDynamic(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="dynList.length === 0" description="暂无动态" />
    </el-dialog>

    <!-- 添加/编辑动态 -->
    <el-dialog v-model="dynEditDialog.visible" :title="dynEditDialog.title" width="550px">
      <el-form label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="dynEditForm.title" placeholder="动态标题" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="dynEditForm.content" type="textarea" :rows="4" placeholder="动态内容" />
        </el-form-item>
        <el-form-item label="图片">
          <div style="display:flex;flex-wrap:wrap;gap:8px;margin-bottom:8px">
            <div v-for="(img, j) in dynEditForm.imageList" :key="j" style="position:relative;width:70px;height:70px">
              <el-image :src="img" style="width:70px;height:70px;border-radius:6px" fit="cover" />
              <div style="position:absolute;top:-6px;right:-6px;width:18px;height:18px;background:rgba(0,0,0,.5);border-radius:50%;display:flex;align-items:center;justify-content:center;cursor:pointer" @click="dynEditForm.imageList.splice(j,1)">✕</div>
            </div>
          </div>
          <el-upload action="/upload" :show-file-list="false" :on-success="handleDynImageSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="image/*">
            <div class="dyn-upload-btn">+</div>
          </el-upload>
        </el-form-item>
        <el-form-item label="视频">
          <div style="display:flex;flex-wrap:wrap;gap:8px;margin-bottom:8px">
            <div v-for="(vurl, j) in dynEditForm.videoList" :key="j" :style="{ position:'relative', width:'70px', height:'70px', borderRadius:'6px', background:`#eee url(${getVideoThumb(vurl)}) center/cover`, display:'flex', alignItems:'center', justifyContent:'center' }">
              <span style="font-size:24px;color:#999">▶</span>
              <div style="position:absolute;top:-6px;right:-6px;width:18px;height:18px;background:rgba(0,0,0,.5);border-radius:50%;display:flex;align-items:center;justify-content:center;cursor:pointer" @click="dynEditForm.videoList.splice(j,1)">✕</div>
            </div>
          </div>
          <el-upload action="/upload" :show-file-list="false" :on-success="handleDynVideoSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="{ Authorization: token }" accept="video/*">
            <div class="dyn-upload-btn">+</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dynEditDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="dynEditSaving" @click="saveDynamicEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 成绩列表 -->
    <el-dialog v-model="scoreDialog.visible" :title="scoreDialog.title" width="720px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div>
          <el-button size="small" type="primary" @click="addScore">添加成绩</el-button>
        </div>
        <div class="toolbar-icons">
          <el-upload :show-file-list="false" :before-upload="importScoresCSV" accept=".csv" style="display:inline-block">
            <el-button circle icon="Upload" title="导入" size="small" />
          </el-upload>
          <el-button circle icon="Download" title="导出 CSV" size="small" @click="exportScores" />
          <el-button circle icon="Sort" title="排序" size="small" @click="ElMessage.info('排序功能开发中')" />
        </div>
      </div>
      <el-table :data="scoreList" v-loading="scoreLoading" stripe>
        <el-table-column label="参赛者" min-width="140">
          <template #default="{ row }">{{ row.participantName || row.participantId }}</template>
        </el-table-column>
        <el-table-column label="部门" min-width="100">
          <template #default="{ row }">{{ row.participantDept || '-' }}</template>
        </el-table-column>
        <el-table-column label="顶级部门" min-width="100">
          <template #default="{ row }">{{ row.participantTopDept || '-' }}</template>
        </el-table-column>
        <el-table-column label="成绩" min-width="200">
          <template #default="{ row }">
            <template v-if="row._parsed && row._parsed.length > 0">
              <span v-for="(ps, j) in row._parsed" :key="j" style="margin-right:8px">{{ ps.name }}:{{ ps.score }}</span>
            </template>
            <span v-else>{{ row.score }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="editScore(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 编辑成绩 -->
    <el-dialog v-model="scoreEditDialog.visible" :title="scoreEditDialog.title" width="500px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="参赛者">
          <el-select v-if="!scoreEditDialog.id" v-model="scoreEditForm.participantId" placeholder="选择参赛者" filterable style="width:100%">
            <el-option v-for="p in scorePartList" :key="p.id" :label="p.userName || p.miniOpenId" :value="p.miniOpenId" />
          </el-select>
          <span v-else>{{ scoreEditForm.participantName }}</span>
        </el-form-item>
        <el-form-item v-for="(sf, j) in scoreEditFields" :key="j" :label="sf.name">
          <el-input v-model="scoreEditForm.scores[j]" :placeholder="'请输入' + sf.name" />
        </el-form-item>
        <el-form-item v-if="scoreEditFields.length === 0" label="成绩">
          <el-input v-model="scoreEditForm.rawScore" placeholder="请输入成绩" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scoreEditDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="scoreEditSaving" @click="saveScoreEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { Delete, ArrowDown } from '@element-plus/icons-vue'
import SortPopover from '../../components/SortPopover.vue'
import { adminApi } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hasPerm } from '../../utils/permission'

const loading = ref(false)
const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const typeFilter = ref<number | ''>('')
const selected = ref<any[]>([])
const saving = ref(false)
const categories = ref<any[]>([])
const deptTree = ref<any[]>([])
const sortRules = ref<{field:string;order:string}[]>([])
const sortColumns = [
  { label: '标题', field: 'title' },
  { label: '类型', field: 'type' },
  { label: '状态', field: 'status' },
  { label: '排序', field: 'order' },
  { label: '已报名', field: 'userCnt' },
  { label: '报名开始', field: 'regStart' },
  { label: '报名结束', field: 'regEnd' },
  { label: '活动开始', field: 'eventStart' },
  { label: '活动结束', field: 'eventEnd' },
  { label: '创建时间', field: 'addTime' },
]
const token = localStorage.getItem('admin_token') || ''

function handleCoverSuccess(res: any) {
  if (res.data?.url) form.cover = res.data.url
}

function fmtDate(ts: number) {
  if (!ts) return '-'
  const d = new Date(ts)
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}

async function load() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value, keyword: keyword.value }
    if (typeFilter.value) params.type = typeFilter.value
    if (sortRules.value.length) params.sort = sortRules.value.map(s => s.field + ':' + s.order).join(',')
    const res = await adminApi.eventList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch { list.value = []; total.value = 0 }
  loading.value = false
}
function onSortChange() { page.value = 1; load() }
function search() { page.value = 1; load() }

const formDialog = reactive({ visible: false, title: '', isCreate: false })
const form = reactive({
  id: null as any, title: '', type: 1, cover: '', desc: '', rules: '', cateName: '', sort: 9999, status: 1,
  regStart: '' as string | number, regEnd: '' as string | number, eventStart: '' as string | number, eventEnd: '' as string | number,
  publishDeptIds: '', fields: [] as any[], scoreFields: [] as any[],
  organizers: [] as any[], assistants: [] as any[], referees: [] as any[]
})

function resetForm() {
  Object.assign(form, {
    id: null, title: '', type: 1, cover: '', desc: '', rules: '', cateName: '', sort: 9999, status: 1,
    regStart: '', regEnd: '', eventStart: '', eventEnd: '',
    publishDeptIds: '', fields: [], scoreFields: [],
    organizers: [], assistants: [], referees: []
  })
}

async function showAdd() {
  resetForm()
  formDialog.isCreate = true
  formDialog.title = '新增赛事活动'
  await loadCategories('activity_type')
  formDialog.visible = true
}

async function showEdit(row: any) {
  resetForm()
  formDialog.isCreate = false
  formDialog.title = '编辑赛事活动'
  try {
    const res = await adminApi.eventDetail(row.id)
    const d = res.data || {}
    form.id = d.id
    form.title = d.title || ''
    form.type = d.type || 1
    form.status = d.status ?? 1
    form.cover = d.img || ''
    form.desc = d.desc || ''
    form.rules = d.rules || ''
    form.sort = d.order ?? 9999
    form.publishDeptIds = d.publishDeptIds || ''
    if (d.regStart) form.regStart = Number(d.regStart)
    if (d.regEnd) form.regEnd = Number(d.regEnd)
    if (d.eventStart) form.eventStart = Number(d.eventStart)
    if (d.eventEnd) form.eventEnd = Number(d.eventEnd)
    if (d.forms) {
      try { form.fields = JSON.parse(d.forms) } catch { form.fields = [] }
    }
    if (d.scoreFields) {
      try { form.scoreFields = JSON.parse(d.scoreFields) } catch { form.scoreFields = [] }
    }
    form.organizers = (d.organizers || []).map((r: any) => ({ userId: r.userId || r.userID, name: r.name }))
    form.assistants = (d.assistants || []).map((r: any) => ({ userId: r.userId || r.userID, name: r.name }))
    form.referees = (d.referees || []).map((r: any) => ({ userId: r.userId || r.userID, name: r.name }))
    await loadCategories(form.type === 2 ? 'competition_type' : 'activity_type')
    form.cateName = d.cateName || ''
  } catch {}
  formDialog.visible = true
}

async function saveEvent() {
  if (!form.title) { ElMessage.warning('请输入标题'); return }
  saving.value = true
  try {
    const payload: any = {
      title: form.title,
      type: form.type,
      status: form.status,
      qr: form.cover || '',
      cateName: form.cateName,
      order: form.sort,
      regStart: form.regStart || '0', regEnd: form.regEnd || '0',
      eventStart: form.eventStart || '0', eventEnd: form.eventEnd || '0',
      forms: JSON.stringify(form.fields),
      scoreFields: JSON.stringify(form.scoreFields),
      obj: JSON.stringify({ desc: form.desc, rules: form.rules, cover: form.cover ? [form.cover] : [] }),
      deptId: 0, publishDeptIds: form.publishDeptIds,
      organizers: JSON.stringify(form.organizers.map((u: any) => u.userId)),
      assistants: JSON.stringify(form.assistants.map((u: any) => u.userId)),
      referees: JSON.stringify(form.referees.map((u: any) => u.userId)),
    }
    if (formDialog.isCreate) {
      await adminApi.eventInsert(payload)
    } else {
      payload.id = form.id
      await adminApi.eventEdit(payload)
    }
    ElMessage.success('保存成功')
    formDialog.visible = false
    load()
  } finally { saving.value = false }
}

// User picker
const userPicker = reactive({ visible: false, role: '', deptId: null as any, users: [] as any[], selected: [] as any[] })
const userPickerTitle = computed(() => {
  const map: any = { organizer: '主办人', assistant: '主办人助理', referee: '裁判' }
  return map[userPicker.role] || ''
})

function showUserPicker(role: string) {
  userPicker.role = role
  userPicker.deptId = null
  userPicker.users = []
  userPicker.selected = []
  userPicker.visible = true
}

async function handleDeptClick(node: any) {
  userPicker.deptId = node.id
  userPicker.selected = []
  if (!node.id) { userPicker.users = []; return }
  try {
    const res = await adminApi.deptUsers({ deptIds: String(node.id) })
    userPicker.users = res.data?.list || []
  } catch { userPicker.users = [] }
}

function confirmUserPicker() {
  const pluralMap: Record<string, string> = { organizer: 'organizers', assistant: 'assistants', referee: 'referees' }
  const key = pluralMap[userPicker.role] || userPicker.role + 's'
  const existing = (form as any)[key] || []
  const existingIds = existing.map((u: any) => u.userId)
  for (const u of userPicker.selected) {
    if (!existingIds.includes(u.openid)) {
      existing.push({ userId: u.openid, name: u.name || u.openid })
    }
  }
  userPicker.visible = false
  ElMessage.success(`已选择 ${userPicker.selected.length} 人`)
}

// Form fields editor
const formEditor = reactive({ visible: false, fields: [] as any[] })

function showFormEditor() {
  formEditor.fields = JSON.parse(JSON.stringify(form.fields || []))
  formEditor.visible = true
}

function confirmFormEditor() {
  form.fields = JSON.parse(JSON.stringify(formEditor.fields))
  formEditor.visible = false
}

// Score fields editor
const scoreFieldEditor = reactive({ visible: false, fields: [] as any[] })

function showScoreFieldEditor() {
  scoreFieldEditor.fields = JSON.parse(JSON.stringify(form.scoreFields || []))
  scoreFieldEditor.visible = true
}

function confirmScoreFieldEditor() {
  form.scoreFields = JSON.parse(JSON.stringify(scoreFieldEditor.fields))
  scoreFieldEditor.visible = false
}

// Participants
const partDialog = reactive({ visible: false, title: '' })
const partList = ref<any[]>([])
const partLoading = ref(false)
const partSelected = ref<any[]>([])
let partEventId = ''

async function showParticipants(row: any) {
  partDialog.title = '参与者 - ' + row.title
  partEventId = row.id
  partDialog.visible = true
  partLoading.value = true
  try {
    const res = await adminApi.eventParticipantList({ eventId: row.id })
    const list = res.data?.list
    partList.value = Array.isArray(list) ? list : []
  } catch { partList.value = [] }
  partLoading.value = false
}

function exportParticipants() {
  const rows = [['用户ID(openid)', '用户昵称', '手机号', '头像', '部门', '顶级部门', '报名时间', '报名IP', '报名表单']]
  partList.value.forEach((p: any) => {
    let formsStr = ''
    if (p.forms) {
      try {
        const parsed = typeof p.forms === 'string' ? JSON.parse(p.forms) : p.forms
        if (Array.isArray(parsed)) {
          formsStr = parsed.map(v => v !== null && v !== undefined ? String(v) : '').join(' | ')
        } else if (typeof parsed === 'object' && parsed) {
          formsStr = Object.entries(parsed).map(([k, v]) => k + ':' + (v ?? '')).join(' | ')
        } else {
          formsStr = String(parsed)
        }
      } catch {
        formsStr = String(p.forms)
      }
    }
    rows.push([
      p.miniOpenId || '',
      p.userName || '',
      p.mobile || '',
      p.userAvatar || '',
      p.deptName || '',
      p.topDeptName || '',
      fmtTime(p._createTime),
      p.EVENT_PART_ADD_IP || p.addIP || '',
      formsStr,
    ])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v ?? '').replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = (partDialog.title.replace('参与者 - ', '') || '报名名单') + '-用户信息.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

async function delParticipant(row: any) {
  try {
    await ElMessageBox.confirm('确定删除该参与者？', '提示')
    await adminApi.eventParticipantDel({ id: row.id })
    ElMessage.success('已删除')
    const idx = partList.value.indexOf(row)
    if (idx > -1) partList.value.splice(idx, 1)
  } catch {}
}

async function delSelectedParticipants() {
  if (partSelected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${partSelected.value.length} 个参与者？`, '提示')
    const ids = partSelected.value.map((r: any) => r.id).join(',')
    await adminApi.eventParticipantDels({ ids })
    ElMessage.success('已删除')
    partSelected.value = []
    // Reload list
    const res = await adminApi.eventParticipantList({ eventId: partEventId })
    const list = res.data?.list
    partList.value = Array.isArray(list) ? list : []
  } catch {}
}

// Dynamics
const dynDialog = reactive({ visible: false, title: '' })
const dynList = ref<any[]>([])
const dynSelected = ref<any[]>([])
let dynEventId = ''

async function showDynamics(row: any) {
  dynDialog.title = '动态 - ' + row.title
  dynEventId = row.id
  try {
    const res = await adminApi.eventDynamics({ eventId: row.id })
    const list = res.data?.list
    dynList.value = Array.isArray(list) ? list : []
  } catch { dynList.value = [] }
  dynDialog.visible = true
}

const dynEditDialog = reactive({ visible: false, title: '' })
const dynEditForm = reactive({ id: '', title: '', content: '', imageList: [] as string[], videoList: [] as string[] })
const dynEditSaving = ref(false)

function addDynamic() {
  dynEditDialog.title = '添加动态'
  dynEditForm.id = ''
  dynEditForm.title = ''
  dynEditForm.content = ''
  dynEditForm.imageList = []
  dynEditForm.videoList = []
  dynEditDialog.visible = true
}

function editDynamic(row: any) {
  dynEditDialog.title = '编辑动态'
  dynEditForm.id = row.id
  dynEditForm.title = row.title || ''
  dynEditForm.content = row.content || ''
  dynEditForm.imageList = row.imageList ? [...row.imageList] : []
  dynEditForm.videoList = row.videoList ? [...row.videoList] : []
  dynEditDialog.visible = true
}

function handleDynImageSuccess(res: any) {
  if (res.code === 0) {
    dynEditForm.imageList.push(res.data.url)
  } else {
    ElMessage.error(res.msg || '上传失败')
  }
}

function handleDynVideoSuccess(res: any) {
  if (res.code === 0) {
    dynEditForm.videoList.push(res.data.url || res.data.thumb)
  } else {
    ElMessage.error(res.msg || '上传失败')
  }
}

function previewVideo(url: string) {
  window.open(url, '_blank')
}

function getVideoThumb(url: string) {
  return url.replace(/\.[^.]+$/, '_thumb.jpg')
}

function onWheel(e: WheelEvent) {
  const el = e.currentTarget as HTMLElement
  if (el.scrollWidth > el.clientWidth) {
    el.scrollLeft += e.deltaY
    e.preventDefault()
  }
}

async function saveDynamicEdit() {
  dynEditSaving.value = true
  try {
    const images = JSON.stringify(dynEditForm.imageList)
    const videos = JSON.stringify(dynEditForm.videoList)
    if (dynEditForm.id) {
      await adminApi.eventDynamicEdit({ id: dynEditForm.id, title: dynEditForm.title, content: dynEditForm.content, images, videos })
    } else {
      await adminApi.eventDynamicAdd({ eventId: dynEventId, title: dynEditForm.title, content: dynEditForm.content, images, videos })
    }
    ElMessage.success('保存成功')
    dynEditDialog.visible = false
    // Reload list
    const res = await adminApi.eventDynamics({ eventId: dynEventId })
    const list = res.data?.list
    dynList.value = Array.isArray(list) ? list : []
  } catch { ElMessage.error('保存失败') }
  dynEditSaving.value = false
}

async function delDynamic(row: any) {
  try {
    await ElMessageBox.confirm('确定删除该动态？', '提示')
    await adminApi.eventDynamicDel({ id: row.id })
    ElMessage.success('已删除')
    const idx = dynList.value.indexOf(row)
    if (idx > -1) dynList.value.splice(idx, 1)
  } catch {}
}

async function delSelectedDynamics() {
  if (dynSelected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${dynSelected.value.length} 条动态？`, '提示')
    const ids = dynSelected.value.map((r: any) => r.id).join(',')
    await adminApi.eventDynamicDels({ ids })
    ElMessage.success('已删除')
    // Reload list
    const res = await adminApi.eventDynamics({ eventId: dynEventId })
    const list = res.data?.list
    dynList.value = Array.isArray(list) ? list : []
    dynSelected.value = []
  } catch {}
}

// Scores
const scoreDialog = reactive({ visible: false, title: '' })
const scoreList = ref<any[]>([])
const scoreLoading = ref(false)
const scorePartList = ref<any[]>([])
let scoreEventId = ''
const scoreEditFields = ref<any[]>([])
const scoreEditDialog = reactive({ visible: false, title: '', id: '' })
const scoreEditForm = reactive({ participantId: '', participantName: '', scores: [] as string[], rawScore: '' })
const scoreEditSaving = ref(false)

async function loadScoreFields(eventId: string) {
  try {
    const res = await adminApi.eventDetail(eventId)
    if (res.data) {
      let sf = res.data.scoreFields
      if (typeof sf === 'string') { try { sf = JSON.parse(sf) } catch { sf = [] } }
      scoreEditFields.value = Array.isArray(sf) ? sf : []
    }
  } catch {}
}

async function showScores(row: any) {
  scoreDialog.title = '成绩 - ' + row.title
  scoreEventId = row.id
  scoreDialog.visible = true
  scoreLoading.value = true
  loadScoreFields(row.id)
  try {
    const res = await adminApi.eventScores({ eventId: row.id })
    const list = res.data?.list || []
    scoreList.value = list.map((s: any) => {
      let _parsed: any[] = []
      try { const t = JSON.parse(s.score); if (Array.isArray(t)) _parsed = t } catch {}
      return { ...s, _parsed }
    })
  } catch { scoreList.value = [] }
  scoreLoading.value = false
}

function addScore() {
  scoreEditDialog.title = '添加成绩'
  scoreEditDialog.id = ''
  scoreEditForm.participantId = ''
  scoreEditForm.participantName = ''
  scoreEditForm.rawScore = ''
  scoreEditForm.scores = scoreEditFields.value.map(() => '')
  // Load participants for dropdown
  adminApi.eventParticipantList({ eventId: scoreEventId }).then(res => {
    scorePartList.value = Array.isArray(res.data?.list) ? res.data.list : []
  }).catch(() => { scorePartList.value = [] })
  scoreEditDialog.visible = true
}

function editScore(row: any) {
  scoreEditDialog.title = '编辑成绩 - ' + (row.participantName || row.participantId)
  scoreEditDialog.id = row.id
  scoreEditForm.participantId = row.participantId
  scoreEditForm.participantName = row.participantName || row.participantId
  scoreEditForm.scores = (row._parsed || []).map((p: any) => p.score)
  scoreEditForm.rawScore = ''
  scoreEditDialog.visible = true
}

async function saveScoreEdit() {
  if (scoreEditFields.value.length > 0) {
    const hasEmpty = scoreEditForm.scores.some(v => !v)
    if (hasEmpty) { ElMessage.warning('请填写所有评分项'); return }
  } else if (!scoreEditForm.rawScore && !scoreEditDialog.id) {
    ElMessage.warning('请填写成绩'); return
  }
  scoreEditSaving.value = true
  try {
    let scoreStr = scoreEditForm.rawScore
    if (scoreEditFields.value.length > 0) {
      scoreStr = JSON.stringify(scoreEditFields.value.map((sf, j) => ({ name: sf.name, score: scoreEditForm.scores[j] })))
    }
    if (scoreEditDialog.id) {
      await adminApi.eventScoreEdit({ id: scoreEditDialog.id, score: scoreStr })
    } else {
      if (!scoreEditForm.participantId) { ElMessage.warning('请选择参赛者'); return }
      await adminApi.eventScoreEdit({ id: '', score: scoreStr, eventId: scoreEventId, participantId: scoreEditForm.participantId })
    }
    ElMessage.success('已保存')
    scoreEditDialog.visible = false
    // Reload
    const res = await adminApi.eventScores({ eventId: scoreEventId })
    const list = res.data?.list || []
    scoreList.value = list.map((s: any) => {
      let _parsed: any[] = []
      try { const t = JSON.parse(s.score); if (Array.isArray(t)) _parsed = t } catch {}
      return { ...s, _parsed }
    })
  } catch { ElMessage.error('保存失败') }
  scoreEditSaving.value = false
}

function importScoresCSV(file: File) {
  const reader = new FileReader()
  reader.onload = async () => {
    const text = reader.result as string
    const lines = text.split('\n').filter(l => l.trim())
    if (lines.length < 2) { ElMessage.warning('CSV 格式错误'); return }
    const header = lines[0].split(',').map(h => h.trim().replace(/^"|"$/g, ''))
    const piIdx = header.findIndex(h => h === 'participantId' || h === '参赛者ID')
    if (piIdx === -1) { ElMessage.warning('CSV 需要 participantId 列'); return }
    let count = 0
    for (let i = 1; i < lines.length; i++) {
      const cols = lines[i].split(',').map(c => c.trim().replace(/^"|"$/g, ''))
      const participantId = cols[piIdx]
      if (!participantId) continue
      let scoreStr = ''
      if (scoreEditFields.value.length > 0) {
        const scores = scoreEditFields.value.map((sf, j) => {
          const colIdx = header.findIndex(h => h === sf.name)
          return { name: sf.name, score: colIdx > -1 ? (cols[colIdx] || '') : '' }
        })
        scoreStr = JSON.stringify(scores)
      } else {
        scoreStr = cols[piIdx + 1] || ''
      }
      try {
        await adminApi.eventScoreEdit({ id: '', eventId: scoreEventId, participantId, score: scoreStr })
        count++
      } catch {}
    }
    ElMessage.success(`成功导入 ${count} 条成绩`)
    // Reload
    const res = await adminApi.eventScores({ eventId: scoreEventId })
    const list = res.data?.list || []
    scoreList.value = list.map((s: any) => {
      let _parsed: any[] = []
      try { const t = JSON.parse(s.score); if (Array.isArray(t)) _parsed = t } catch {}
      return { ...s, _parsed }
    })
  }
  reader.readAsText(file)
  return false // prevent upload
}

function exportScores() {
  const rows = [['参赛者', '部门', '顶级部门', '成绩']]
  scoreList.value.forEach((s: any) => {
    let scoreStr = s.score
    if (s._parsed && s._parsed.length > 0) {
      scoreStr = s._parsed.map((ps: any) => `${ps.name}:${ps.score}`).join('; ')
    }
    rows.push([
      s.participantName || s.participantId || '',
      s.participantDept || '',
      s.participantTopDept || '',
      scoreStr,
    ])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = (scoreDialog.title.replace('成绩 - ', '') || '成绩') + '.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

// More actions
async function handleMore(cmd: string, row: any) {
  if (cmd === 'enable') {
    await adminApi.eventStatus({ id: row.id, status: '1' })
    ElMessage.success('已启用'); load()
  } else if (cmd === 'disable') {
    await adminApi.eventStatus({ id: row.id, status: '0' })
    ElMessage.success('已停用'); load()
  } else if (cmd === 'del') {
    try {
      await ElMessageBox.confirm('确定删除？', '提示')
      await adminApi.eventDel({ id: row.id })
      ElMessage.success('已删除'); load()
    } catch {}
  } else if (cmd === 'vouch') {
    await adminApi.eventVouch({ id: row.id, vouch: '1' })
    ElMessage.success('已推荐到首页'); load()
  } else if (cmd === 'unvouch') {
    await adminApi.eventVouch({ id: row.id, vouch: '0' })
    ElMessage.success('已取消推荐'); load()
  } else if (cmd === 'top') {
    await adminApi.eventTop({ id: row.id, top: '1' })
    ElMessage.success('已置顶'); load()
  } else if (cmd === 'untop') {
    await adminApi.eventTop({ id: row.id, top: '0' })
    ElMessage.success('已取消置顶'); load()
  } else if (cmd === 'scores') {
    showScores(row)
  }
}

async function delSelected() {
  if (selected.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selected.value.length} 个赛事活动？`, '提示')
    const ids = selected.value.map((r: any) => r.id).join(',')
    await adminApi.eventDels({ ids })
    ElMessage.success('已删除')
    selected.value = []
    load()
  } catch {}
}

function exportData() {
  const rows = [['标题', '类型', '状态', '报名开始', '报名结束', '活动开始', '活动结束', '参与人数']]
  list.value.forEach((r: any) => {
    rows.push([r.title, r.type === 1 ? '活动' : '赛事', r.status === 1 ? '正常' : '停用', fmtTime(r.regStart), fmtTime(r.regEnd), fmtTime(r.eventStart), fmtTime(r.eventEnd), String(r.userCnt || 0)])
  })
  const csv = '\uFEFF' + rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = '赛事活动列表.csv'
  a.click()
  URL.revokeObjectURL(a.href)
}

function getDeptPath(id: number): string {
  const walk = (nodes: any[], parents: string[]): string => {
    for (const n of nodes) {
      const path = [...parents, n.name]
      if (n.id === id) return path.join('/')
      if (n.children) { const r = walk(n.children, path); if (r) return r }
    }
    return ''
  }
  return walk(deptTree.value, []) || ''
}

function getDeptNames(ids: string): string {
  if (!ids) return '选择部门'
  const names = ids.split(',').map(id => getDeptPath(Number(id))).filter(Boolean)
  return names.length ? names.join(', ') : '选择部门'
}

function fmtTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function loadCategories(dictKey: string) {
  try {
    const res = await adminApi.dictItems(dictKey)
    categories.value = (res.data || []).map((d: any) => ({ label: d.label, value: d.value }))
  } catch { categories.value = [] }
}

watch(() => form.type, (t) => {
  form.cateName = ''
  loadCategories(t === 2 ? 'competition_type' : 'activity_type')
})

async function loadDeptTree() {
  try {
    const res = await adminApi.deptTree()
    deptTree.value = res.data || []
  } catch { deptTree.value = [] }
}

onMounted(() => { load(); loadCategories('activity_type'); loadDeptTree() })
</script>

<style scoped>
.dept-tree-wrap {
  width: 220px;
  flex-shrink: 0;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: auto;
  max-height: 400px;
}
.dept-users-wrap {
  flex: 1;
  min-width: 0;
}
.dept-tree-title {
  padding: 8px 12px;
  font-weight: bold;
  font-size: 13px;
  color: #666;
  border-bottom: 1px solid #ebeef5;
  background: #f5f7fa;
}
.multi-select-input {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  min-height: 32px;
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 2px 8px;
  cursor: pointer;
  transition: border-color 0.2s;
  background: #fff;
  box-sizing: border-box;
}
.multi-select-input:hover {
  border-color: #409eff;
}
.ms-placeholder {
  color: #c0c4cc;
  font-size: 14px;
  flex: 1;
}
.ms-arrow {
  color: #c0c4cc;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: auto;
}
.cover-upload {
  position: relative;
  width: 100px;
  height: 100px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.3s;
}
.cover-upload:hover {
  border-color: #409eff;
}
.cover-placeholder {
  font-size: 32px;
  color: #999;
  line-height: 100px;
  text-align: center;
}
.cover-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-overlay {
  position: absolute;
  top: 0;
  right: 0;
  padding: 4px;
}
.scroll-x {
  display: flex;
  gap: 4px;
  flex-wrap: nowrap;
  overflow-x: auto;
  white-space: nowrap;
  padding-bottom: 4px;
  scrollbar-width: none;
  -ms-overflow-style: none;
}
.scroll-x::-webkit-scrollbar { display: none; }
.dyn-upload-btn {
  width: 70px;
  height: 70px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #999;
  transition: border-color 0.3s;
}
.dyn-upload-btn:hover {
  border-color: #409eff;
  color: #409eff;
}
.toolbar-icons {
  display: flex;
  align-items: center;
}
.toolbar-icons > * {
  margin-left: 8px;
}
.toolbar-icons > :first-child {
  margin-left: 0;
}
</style>
<style>
.el-tree { text-align: left; }
</style>

package com.wecheckin.app.ddauth

import android.app.Activity
import android.content.Context
import android.os.Bundle
import com.android.dingtalk.openauth.utils.DDAuthConstant
import org.json.JSONObject

class DDAuthCallbackActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val result = JSONObject()
            .put("authCode", intent?.getStringExtra(DDAuthConstant.CALLBACK_EXTRA_AUTH_CODE).orEmpty())
            .put("state", intent?.getStringExtra(DDAuthConstant.CALLBACK_EXTRA_STATE).orEmpty())
            .put("error", intent?.getStringExtra(DDAuthConstant.CALLBACK_EXTRA_ERROR).orEmpty())

        getSharedPreferences(PREFERENCES_NAME, Context.MODE_PRIVATE)
            .edit()
            .putString(RESULT_KEY, result.toString())
            .commit()
        finish()
    }

    companion object {
        private const val PREFERENCES_NAME = "wecheckin_dingtalk_native_auth"
        private const val RESULT_KEY = "result"
    }
}

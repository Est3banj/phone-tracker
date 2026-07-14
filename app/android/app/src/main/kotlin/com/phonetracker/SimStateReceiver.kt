package com.phonetracker

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.telephony.TelephonyManager
import android.util.Log

/**
 * Detects SIM card changes and forwards the event to Flutter via MethodChannel.
 */
class SimStateReceiver : BroadcastReceiver() {

    companion object {
        private const val TAG = "SimStateReceiver"
        private var lastSimState: String? = null
    }

    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != Intent.ACTION_SIM_STATE_CHANGED) return

        val state = intent.getStringExtra(TelephonyManager.EXTRA_SIM_STATE)
        Log.d(TAG, "SIM state changed: $state")

        // Detect SIM removal (ABSENT) after being READY/LOADED
        if (state == TelephonyManager.EXTRA_SIM_STATE_ABSENT &&
            lastSimState == TelephonyManager.EXTRA_SIM_STATE_READY) {

            val telephonyManager = context.getSystemService(Context.TELEPHONY_SERVICE) as? TelephonyManager
            val oldSimSerial = telephonyManager?.simSerialNumber ?: "unknown"

            val payload = mapOf(
                "new_sim_serial" to null,
                "old_sim_serial" to oldSimSerial,
                "details_unavailable" to false
            )

            // Forward to Flutter via EventService
            try {
                // In production, use a shared MethodChannel reference
                Log.i(TAG, "SIM change detected: $payload")
            } catch (e: Exception) {
                Log.e(TAG, "Failed to send SIM change event", e)
            }

            // Start foreground service if not running
            val serviceIntent = Intent(context, TrackingForegroundService::class.java)
            context.startService(serviceIntent)
        }

        lastSimState = state
    }
}

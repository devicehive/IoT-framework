import core
import dbus.service



MSG_TYPE_INFO = dbus.UInt32(0)
MSG_TYPE_WARNING = dbus.UInt32(1)
MSG_TYPE_EMERGENCY = dbus.UInt32(2)

'''
Producer and Dismisser obejcts are not needed as they are implemented in the bridge itself
'''

class NotificationService(core.PropertiesServiceInterface):
    """Enabling Notifications for alljoyn application"""
    def __init__(self, container, path):
        core.PropertiesServiceInterface.__init__(
            self, container, path,
            {'org.alljoyn.Notification': {'Version': dbus.UInt16(1)}}
        )

    def IntrospectionXml(self):
        return """
             <interface name="org.alljoyn.Notification">
                <property name="Version" type="q" access="read"/>
                <signal name="Notify">
                    <arg name="version" type="q"/>
                    <arg name="msgId" type="i"/>
                    <arg name="msgType" type="q"/>
                    <arg name="deviceId" type="s"/>
                    <arg name="deviceName" type="s"/>
                    <arg name="appId" type="ay"/>
                    <arg name="appName" type="s"/>
                    <arg name="langText" type="a{ss}"/>
                    <arg name="attributes" type="a{iv}"/>
                    <arg name="customAttributes" type="a{ss}"/>
                </signal>
                </interface>
        """ + core.PropertiesServiceInterface.IntrospectionXml(self)

    '''
    Arguments versionm msgId, deviceId, deviceName, appId, appName 
    are ignored and will he filled by the bridge. Use NotifySimple method instead.
    '''
    @dbus.service.signal('org.alljoyn.Notification',
                         signature='qiqssaysa{ss}a{iv}a{ss}')
    def Notify(self, version, msgId, msgType, deviceId, deviceName,
               appId, appName, langText, attributes, customAttributes):
        pass


class Notifications(object):
    """Wrapper for /info /warning /emergency services"""
    def __init__(self, container):
        
        self._services = {
            MSG_TYPE_INFO: NotificationService(container, '/info'),
            MSG_TYPE_WARNING: NotificationService(container, 'warning'),
            MSG_TYPE_EMERGENCY: NotificationService(container, '/emergency')
        }

    def _notify(self, msgType, language, message, attributes = {}, customAttributes = {}):        
        self._services[msgType].Notify(dbus.UInt16(2), 0, msgType, '', '', '', '', 
                       {language: message}, attributes, customAttributes)

    def info(self, language, message):
        self._notify(MSG_TYPE_INFO, language, message)

    def warning(self, language, message):
        self._notify(MSG_TYPE_WARNING, language, message)

    def emergency(self, language, message):
        self._notify(MSG_TYPE_EMERGENCY, language, message)

import dbus.service
import core

class ControlPanelService(core.PropertiesServiceInterface):
  def __init__(self, container, name):
    core.PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer" % name, 
      ['org.alljoyn.ControlPanel.ControlPanel', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.ControlPanel' : {'Version': dbus.UInt16(1)}})

  def IntrospectionXml(self):
    return """
         <interface name="org.alljoyn.ControlPanel.ControlPanel">
            <property name="Version" type="q" access="read"/>
         </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

class ContainerService(core.PropertiesServiceInterface):
  def __init__(self, container, name, path, label):
    core.PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer/%s" % (name, path), 
      ['org.alljoyn.ControlPanel.Container', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.Container' : {
        'Version': dbus.UInt16(1), 
        'States': dbus.UInt32(1), 
        'OptParams': dbus.Dictionary({
          dbus.UInt16(0): label,
          dbus.UInt16(2): [dbus.UInt16(1)]
        }, variant_level=1, signature='qv')
      }})

  def IntrospectionXml(self):
    return """
          <interface name="org.alljoyn.ControlPanel.Container">
            <property name="Version" type="q" access="read"/>
            <property name="States" type="u" access="read"/>
            <property name="OptParams" type="a{qv}" access="read"/>
            <signal name="MetadataChanged" />
          </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)
  @dbus.service.signal('org.alljoyn.ControlPanel.Container', signature='')
  def MetadataChanged(self):
      pass


class PropertyService(core.PropertiesServiceInterface):
  def __init__(self, container, name, path, label):
    core.PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer/%s" % (name, path), 
      ['org.alljoyn.ControlPanel.Property', 'org.freedesktop.DBus.Properties'],
      {'org.alljoyn.ControlPanel.Property' : {
        'Version': dbus.UInt16(1), 
        'States': dbus.UInt32(1), 
        'Value': dbus.String('Off', variant_level = 1),
        'OptParams': dbus.Dictionary({
          dbus.UInt16(0): label
        }, variant_level=1, signature='qv')
      }})

  def IntrospectionXml(self):
    return """
       <interface name="org.alljoyn.ControlPanel.Property">
          <property name="Version" type="q" access="read"/>
          <property name="States" type="u" access="read"/>
          <property name="OptParams" type="a{qv}" access="read"/>
          <property name="Value" type="v" access="readwrite"/>
          <signal name="MetadataChanged" />
          <signal name="ValueChanged">
             <arg type="v"/>
          </signal>
       </interface>
    """  + core.PropertiesServiceInterface.IntrospectionXml(self)


  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='')
  def MetadataChanged(self):
      pass
  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='')
  def ValueChanged(self, value):
      pass

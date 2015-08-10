import dbus.service
import core

class ControlPanelService(core.PropertiesServiceInterface):
  def __init__(self, container, unit):
    core.PropertiesServiceInterface.__init__(self, container, "/ControlPanel/%s/rootContainer" % unit, 
      {'org.alljoyn.ControlPanel.ControlPanel' : {'Version': dbus.UInt16(1)}})

  def IntrospectionXml(self):
    return """
         <interface name="org.alljoyn.ControlPanel.ControlPanel">
            <property name="Version" type="q" access="read"/>
         </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

class HttpControlService(core.PropertiesServiceInterface):
  def __init__(self, container, unit, url):
    core.PropertiesServiceInterface.__init__(self, container, "/Control/%s/HTTPControl" % unit, 
      {'org.alljoyn.ControlPanel.HTTPControl' : {'Version': dbus.UInt16(1)}})
    self._url = url

  def IntrospectionXml(self):
    return """
        <interface name="org.alljoyn.ControlPanel.HTTPControl">
          <property name="Version" type="q" access="read"/>
          <method name="GetRootURL">
             <arg name="url" type="s" direction="out"/>
          </method>
       </interface>
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

  @dbus.service.method('org.alljoyn.ControlPanel.HTTPControl', in_signature='', out_signature='s')
  def GetRootURL(self):
    return self._url

WIDGET_STATE_DISABLED = dbus.UInt32(0)
WIDGET_STATE_ENABLED = dbus.UInt32(1)
WIDGET_STATE_WRITABLE = dbus.UInt32(2)

WIDGET_METADATA_LABEL = dbus.UInt16(0)
WIDGET_METADATA_BGCOLOR = dbus.UInt16(1)
CONTAINER_METADATA_LAYOUT_HINTS = dbus.UInt16(2)
CONTAINER_LAYOUT_VERTICAL = dbus.UInt16(1)
CONTAINER_LAYOUT_HORIZONTAL = dbus.UInt16(2)

class BaseWidgetService(core.PropertiesServiceInterface):
  def __init__(self, container, path, interface):
    self._interface = interface
    self._optParams = dbus.Dictionary({WIDGET_METADATA_LABEL: ""}, variant_level=1, signature='qv')

    core.PropertiesServiceInterface.__init__(self, container, path, 
      {interface : {
        'Version': dbus.UInt16(1), 
        'States': WIDGET_STATE_ENABLED, 
        'OptParams': self._optParams
      }})

  def IntrospectionXml(self):
    return core.PropertiesServiceInterface.IntrospectionXml(self)

  def MetadataChanged(self):
    pass    

  def SetStates(self, states):
    self.Set(self._interface, 'States', states)
    self.MetadataChanged()

  def SetOptParam(self, param, value):
    self._optParams[param] = value
    self.MetadataChanged()


class ContainerService(BaseWidgetService):
  def __init__(self, container, path):
    BaseWidgetService.__init__(self, container, path, 
      'org.alljoyn.ControlPanel.Container')

  def IntrospectionXml(self):
    return """
          <interface name="org.alljoyn.ControlPanel.Container">
            <property name="Version" type="q" access="read"/>
            <property name="States" type="u" access="read"/>
            <property name="OptParams" type="a{qv}" access="read"/>
            <signal name="MetadataChanged" />
          </interface>
    """ + BaseWidgetService.IntrospectionXml(self)
  
  @dbus.service.signal('org.alljoyn.ControlPanel.Container', signature='')
  def MetadataChanged(self):
      pass


PROPERTY_METADATA_HINTS = dbus.UInt16(2)
PROPERTY_METADATA_UNITS = dbus.UInt16(3)
PROPERTY_METADATA_CONSTRAIN = dbus.UInt16(4)
PROPERTY_METADATA_RANGE = dbus.UInt16(5)

PROPERTY_WIDGET_HINT_SWITCH         = dbus.UInt16(1)
PROPERTY_WIDGET_HINT_CHECKBOX       = dbus.UInt16(2)
PROPERTY_WIDGET_HINT_SPINNER        = dbus.UInt16(3)
PROPERTY_WIDGET_HINT_RADIOBUTTON    = dbus.UInt16(4)
PROPERTY_WIDGET_HINT_SLIDER         = dbus.UInt16(5)
PROPERTY_WIDGET_HINT_TIMEPICKER     = dbus.UInt16(6)
PROPERTY_WIDGET_HINT_DATEPICKER     = dbus.UInt16(7)
PROPERTY_WIDGET_HINT_NUMBERPICKER   = dbus.UInt16(8)
PROPERTY_WIDGET_HINT_NUMERICKEYPAD  = dbus.UInt16(9)
PROPERTY_WIDGET_HINT_ROTARYKNOB     = dbus.UInt16(10)
PROPERTY_WIDGET_HINT_TEXTLABEL      = dbus.UInt16(11)
PROPERTY_WIDGET_HINT_NUMERICVIEW    = dbus.UInt16(12)
PROPERTY_WIDGET_HINT_EDITTEXT       = dbus.UInt16(13)


class PropertyService(BaseWidgetService):
  def __init__(self, container, path, readonly = False):
    BaseWidgetService.__init__(self, container, path, 'org.alljoyn.ControlPanel.Property')

    if not readonly: 
      self.SetStates(WIDGET_STATE_WRITABLE)


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
    """  + BaseWidgetService.IntrospectionXml(self)

  def SetValue(self, value):
    self.Set(self._interface, 'Value', value)
    self.ValueChanged(value)    

  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='')
  def MetadataChanged(self):
    print("MetadataChanged signal")
    pass

  @dbus.service.signal('org.alljoyn.ControlPanel.Property', signature='v')
  def ValueChanged(self, value):    
    print("ValueChanged signal: %s" % value)
    pass


LABEL_METADATA_HINTS = dbus.UInt16(2)
LABEL_WIDGET_HINT_TEXTLABEL = dbus.UInt16(1)

class LabelService(BaseWidgetService):
  def __init__(self, container, path, label):
    BaseWidgetService.__init__(self, container, path, 'org.alljoyn.ControlPanel.LabelProperty')
    self.Set(self._interface, 'Label', label)
    self.SetOptParam(LABEL_METADATA_HINTS, [LABEL_WIDGET_HINT_TEXTLABEL])

  def IntrospectionXml(self):
    return """
       <interface name="org.alljoyn.ControlPanel.LabelProperty">
          <property name="Version" type="q" access="read"/>
          <property name="States" type="u" access="read"/>
          <property name="Label" type="s" access="read"/>
          <property name="OptParams" type="a{qv}" access="read"/>
          <signal name="MetadataChanged" />
       </interface>
    """  + BaseWidgetService.IntrospectionXml(self)

  @dbus.service.signal('org.alljoyn.ControlPanel.LabelProperty', signature='')
  def MetadataChanged(self):
      pass


ACTION_METADATA_HINTS = dbus.UInt16(2)
ACTION_WIDGET_HINT_ACTIONBUTTON = dbus.UInt16(1)

class ActionService(BaseWidgetService):
  def __init__(self, container, path, exechandler = None):
    BaseWidgetService.__init__(self, container, path, 'org.alljoyn.ControlPanel.Action')
    self.SetOptParam(ACTION_METADATA_HINTS, [ACTION_WIDGET_HINT_ACTIONBUTTON])
    self._exechandler = exechandler


  def SetHandler(self, exechandler):
    self._exechandler = exechandler    

  def IntrospectionXml(self):
    return """
        <interface name="org.alljoyn.ControlPanel.Action">
          <property name="Version" type="q" access="read"/>
          <property name="States" type="u" access="read"/>
          <property name="OptParams" type="a{qv}" access="read"/>
          <signal name="MetadataChanged" />
          <method name="Exec"/>
       </interface>
    """  + BaseWidgetService.IntrospectionXml(self)

  @dbus.service.signal('org.alljoyn.ControlPanel.Action', signature='')
  def MetadataChanged(self):
      pass

  @dbus.service.method('org.alljoyn.ControlPanel.Action', in_signature='', out_signature='')
  def Exec(self):
      print("Exec is called")
      if self._exechandler is not None:
        self._exechandler()      

import dbus.service

ABOUT_IFACE = 'org.alljoyn.About'
CONFIG_SERVICE_IFACE = 'org.alljoyn.Config'

INTROSPECTION_TEMPLATE = """
      <node name="{0}" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:noNamespaceSchemaLocation="http://www.allseenalliance.org/schemas/introspect.xsd">
         {1}
         <interface name="org.freedesktop.DBus.Introspectable"><method name="Introspect"><arg name="out" type="s" direction="out"></arg></method></interface>
         {2}
      </node>
    """

PROPERTIES_IFACE_XML = """
      <interface name="org.freedesktop.DBus.Properties">
        <method name="Get">
          <arg direction="in" name="interface" type="s"/>
          <arg direction="in" name="propname" type="s"/>
          <arg direction="out" name="value" type="v"/>
        </method>
        <method name="GetAll">
          <arg direction="in" name="interface" type="s"/>
          <arg direction="out" name="props" type="a{sv}"/>
        </method>
        <method name="Set">
          <arg direction="in" name="interface" type="s"/>
          <arg direction="in" name="propname" type="s"/>
          <arg direction="in" name="value" type="v"/>
        </method>
      </interface>
    """


def flatten(d, parent_key='', sep='.'):
    items = []
    for k, v in d.items():
        new_key = parent_key + sep + k if parent_key else k
        try:
            items.extend(flatten(v, new_key, sep=sep).items())
        except:
            items.append((new_key, v))
    return dict(items)


class BusContainer(object):
  def __init__(self, bus, root):
    self._bus = bus
    self._root = root.rstrip('/')

  @property
  def bus(self):
    return self._bus

  def relative(self, path):
    if path is None:
      return None
    else:
      return "%s/%s" % (self._root, path.lstrip('/'))

class ServiceInterface(dbus.service.Object):
  def __init__(self, container, path):
    self._path = path    
    if type(path) is list:
      dbus.service.Object.__init__(self, container.bus, None)
      self.SUPPORTS_MULTIPLE_OBJECT_PATHS = True
      for itm in path:
        self.add_to_connection(container.bus.get_bus() , container.relative(itm))
    else:
      dbus.service.Object.__init__(self, container.bus, container.relative(path))

  @property
  def path(self):
      return self._path

  def relative(self, path):
    return "%s/%s" % (self._path, path.lstrip('/'))

  def IntrospectionXml(self):
    return None

  def _selectPath(self, object_path):
    if type(self.path) is list:
      return next(p for p in self.path if object_path.endswith(p))
    else:
      return self.path

  def Introspect(self, object_path, connection):
    xml = self.IntrospectionXml()
    if xml is None:
      return dbus.service.Object.Introspect(self, object_path, connection)
    else:      

      children = "\n".join(['<node name="%s"/>' % name for name in connection.list_exported_child_objects(object_path)])
      return INTROSPECTION_TEMPLATE.format(self._selectPath(object_path), xml, children)


  def introspect(self):    
    introspection = self.Introspect(self._object_path, self._connection)
    # print(introspection)
    return introspection    

  @property
  def object_path(self):
      return self._object_path
  

class PropertiesServiceInterface(ServiceInterface):
  def __init__(self, container, path, properties):
    self._path = path
    self._properties = flatten(properties)
    ServiceInterface.__init__(self, container, path)

  ## dbus.PROPERTIES_IFACE
  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ss', out_signature='v')
  def Get(self, interface, property):
      prop = interface + '.' + property
      print("Properties.Get is called %s" % prop)
      if prop in self._properties:
          return self._properties[prop]
      else:
          raise Exception('Unsupported property: %s' % prop)

  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='ssv')
  def Set(self, interface, property, value):
      prop = interface + '.' + property
      print("Properties.Set is called %s with %s (%s:%d)" % (prop, value, type(value), value.variant_level))
      self._properties[prop] = value
      
  @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature='s', out_signature='a{sv}')
  def GetAll(self, interface):
      print("Properties.GetAll is called for %s" % interface)
      prefix = interface + '.'
      return  {k[len(prefix):]: v for k, v in self._properties.items() if k.startswith(prefix)}

  def IntrospectionXml(self):
    return PROPERTIES_IFACE_XML


class AboutService(PropertiesServiceInterface):
  def __init__(self, container, properties):
    PropertiesServiceInterface.__init__(self, container, '/About',       
      {'org.alljoyn.About' : properties})

  ## org.alljoyn.About Interface

  @dbus.service.method(ABOUT_IFACE, in_signature='s', out_signature='a{sv}')
  def GetAboutData(self, languageTag):
      print("GetAboutData is called")
      return PropertiesServiceInterface.GetAll(self, ABOUT_IFACE)

  @dbus.service.method(ABOUT_IFACE, in_signature='', out_signature='a(oas)')
  def GetObjectDescription(self):
      print('GetObjectDescription - empty')
      return {}

  @dbus.service.signal(ABOUT_IFACE, signature='qqa(oas)a{sv}')
  def Announce(self, version, port, objectDescription, metaData):
      print('Announce - empty')
      pass

  def IntrospectionXml(self):
    return """
         <interface name="org.alljoyn.About">
            <property name="Version" type="q" access="read"/>
            <method name="GetAboutData">
               <arg name="languageTag" type="s" direction="in"/>
               <arg name="aboutData" type="a{sv}" direction="out"/>
            </method>
            <method name="GetObjectDescription">
               <arg name="objectDescription" type="a(oas)" direction="out"/>
            </method>
            <signal name="Announce">
               <arg name="version" type="q"/>
               <arg name="port" type="q"/>
               <arg name="objectDescription" type="a(oas)"/>
               <arg name="metaData" type="a{sv}"/>
            </signal>
         </interface>
    """  + PropertiesServiceInterface.IntrospectionXml(self)

class ConfigService(PropertiesServiceInterface):
  def __init__(self, container, devicename):
    self._devicename = devicename
    PropertiesServiceInterface.__init__(self, container, '/Config', 
      {'org.alljoyn.Config' : {'Version': 1}})

  ## org.alljoyn.Config Interface
  @dbus.service.method(CONFIG_SERVICE_IFACE, in_signature='s', out_signature='a{sv}')
  def GetConfigurations(self, languageTag):
      print('GetConfigurations: %s', self._devicename)
      
      return {            
          'DefaultLanguage': 'en',
          'DeviceName': self._devicename
      }

  def IntrospectionXml(self):
    return """
       <interface name="org.alljoyn.Config">
          <property name="Version" type="q" access="read"/>
          <method name="FactoryReset">
             <annotation name="org.freedesktop.DBus.Method.NoReply" value="true"/>
          </method>
          <method name="Restart">
             <annotation name="org.freedesktop.DBus.Method.NoReply" value="true"/>
          </method>
          <method name="SetPasscode">
             <arg name="daemonRealm" type="s" direction="in"/>
             <arg name="newPasscode" type="ay" direction="in"/>
          </method>
          <method name="GetConfigurations">
             <arg name="languageTag" type="s" direction="in"/>
             <arg name="configData" type="a{sv}" direction="out"/>
          </method>
          <method name="UpdateConfigurations">
             <arg name="languageTag" type="s" direction="in"/>
             <arg name="configMap" type="a{sv}" direction="in"/>
          </method>
          <method name="ResetConfigurations">
             <arg name="languageTag" type="s" direction="in"/>
             <arg name="fieldList" type="as" direction="in"/>
          </method>
       </interface>
    """ + PropertiesServiceInterface.IntrospectionXml(self)


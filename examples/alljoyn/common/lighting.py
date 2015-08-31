import dbus.service
import core


LAMP_SERVICE_IFACE = 'org.allseen.LSF.LampService'
LAMP_PARAMETERS_IFACE = 'org.allseen.LSF.LampParameters'
LAMP_DETAILS_IFACE = 'org.allseen.LSF.LampDetails'
LAMP_STATE_IFACE = 'org.allseen.LSF.LampState'


LAMP_OK = dbus.UInt32(0)
LAMP_ERR_NULL = dbus.UInt32(1)
LAMP_ERR_UNEXPECTED = dbus.UInt32(2)

LAMP_OK                          = dbus.UInt32(0),  # Success status 
LAMP_ERR_NULL                    = dbus.UInt32(1),  # Unexpected NULL pointer
LAMP_ERR_UNEXPECTED              = dbus.UInt32(2),  # An operation was unexpected at this time 
LAMP_ERR_INVALID                 = dbus.UInt32(3),  # A value was invalid 
LAMP_ERR_UNKNOWN                 = dbus.UInt32(4),  # A unknown value 
LAMP_ERR_FAILURE                 = dbus.UInt32(5),  # A failure has occurred 
LAMP_ERR_BUSY                    = dbus.UInt32(6),  # An operation failed and should be retried later 
LAMP_ERR_REJECTED                = dbus.UInt32(7),  # The connection was rejected 
LAMP_ERR_RANGE                   = dbus.UInt32(8),  # Value provided was out of range 
LAMP_ERR_INVALID_FIELD           = dbus.UInt32(10), # Invalid param/state field 
LAMP_ERR_MESSAGE                 = dbus.UInt32(11), # An invalid message was received 
LAMP_ERR_INVALID_ARGS            = dbus.UInt32(12), # The arguments were invalid 
LAMP_ERR_EMPTY_NAME              = dbus.UInt32(13), # The name was empty 
LAMP_ERR_RESOURCES               = dbus.UInt32(14), # Out of memory
LAMP_ERR_REPLY_WITH_INVALID_ARGS = dbus.UInt32(15) # The reply received for a message had invalid arguments 


class LampService(core.PropertiesServiceInterface):
  def __init__(self, container, name):
    core.PropertiesServiceInterface.__init__(self, container, "/org/allseen/LSF/Lamp", 
      {LAMP_SERVICE_IFACE : {
        'Version': dbus.UInt32(1),
        'LampServiceVersion': dbus.UInt32(1),
        'LampFaults': []
      },
      LAMP_PARAMETERS_IFACE : {
        'Version': dbus.UInt32(1),
        'Energy_Usage_Milliwatts': dbus.UInt32(0),
        'Brightness_Lumens': dbus.UInt32(0),
      },
      LAMP_DETAILS_IFACE : {
        'Version': dbus.UInt32(1),
        'Make': dbus.UInt32(1),
        'Model': dbus.UInt32(1),
        'Type': dbus.UInt32(1),
        'LampType': dbus.UInt32(30),
        'LampBaseType': dbus.UInt32(8),
        'LampBeamAngle': dbus.UInt32(100),
        'Dimmable': dbus.Boolean(0),
        'Color': dbus.Boolean(0), 
        'VariableColorTemp': dbus.Boolean(0),
        'HasEffects': dbus.Boolean(0),
        'MinVoltage': dbus.UInt32(90),
        'MaxVoltage': dbus.UInt32(240),
        'Wattage': dbus.UInt32(9),
        'IncandescentEquivalent': dbus.UInt32(75),
        'MaxLumens': dbus.UInt32(900),
        'MinTemperature': dbus.UInt32(2700),
        'MaxTemperature': dbus.UInt32(5400),
        'ColorRenderingIndex': dbus.UInt32(0),
        'LampID': name
      },
      LAMP_STATE_IFACE : {
        'Version': dbus.UInt32(1),
         ## In percentage, 0 means 0. uint32_max-1 means 100. */
        'Hue': dbus.UInt32(99999), 
        'Saturation': dbus.UInt32(99999),
        'ColorTemp': dbus.UInt32(99999),
        'Brightness': dbus.UInt32(100),
        'OnOff': dbus.Boolean(0)
      }
      })

  @dbus.service.method(LAMP_SERVICE_IFACE, in_signature='u', out_signature='uu')
  def ClearLampFault(self, LampFaultCode):
    print('ClearLampFault - %s' % LampFaultCode)
    return [LAMP_ERR_UNEXPECTED, LampFaultCode]

    ## org.allseen.LSF.LampState Interface

  @dbus.service.signal(LAMP_STATE_IFACE, signature='s')
  def LampStateChanged(self, LampID):
      print('LampStateChanged - empty')
      pass


  @dbus.service.method(LAMP_STATE_IFACE, in_signature='ta{sv}u', out_signature='u')
  def TransitionLampState(self, Timestamp, NewState, TransitionPeriod):
      print('TransitionLampState  %s' % NewState['OnOff'])
      self.Set(LAMP_STATE_IFACE, 'OnOff', NewState['OnOff'])
      return 1


  @dbus.service.method(LAMP_STATE_IFACE, in_signature='a{sv}a{sv}uuut', out_signature='u')
  def ApplyPulseEffect(self, FromState, ToState, period, duration, numPulses, timestamp):
      print('ApplyPulseEffect - empty')
      return 0



  def IntrospectionXml(self):
    return """
         <interface name="org.allseen.LSF.LampService">
          <property name="Version" type="u" access="read"/>
          <property name="LampServiceVersion" type="u" access="read"/>
          <method name="ClearLampFault">
            <arg name="LampFaultCode" type="u" direction="in"/>
            <arg name="LampResponseCode" type="u" direction="out"/>
            <arg name="LampFaultCode" type="u" direction="out"/>
          </method>
          <property name="LampFaults" type="au" access="read"/>
        </interface>
        <interface name="org.allseen.LSF.LampParameters">
          <property name="Version" type="u" access="read"/>
          <property name="Energy_Usage_Milliwatts" type="u" access="read"/>
          <property name="Brightness_Lumens" type="u" access="read"/>
        </interface>
        <interface name="org.allseen.LSF.LampDetails">
          <property name="Version" type="u" access="read"/>
          <property name="Make" type="u" access="read"/>
          <property name="Model" type="u" access="read"/>
          <property name="Type" type="u" access="read"/>
          <property name="LampType" type="u" access="read"/>
          <property name="LampBaseType" type="u" access="read"/>
          <property name="LampBeamAngle" type="u" access="read"/>
          <property name="Dimmable" type="b" access="read"/>
          <property name="Color" type="b" access="read"/>
          <property name="VariableColorTemp" type="b" access="read"/>
          <property name="HasEffects" type="b" access="read"/>
          <property name="MinVoltage" type="u" access="read"/>
          <property name="MaxVoltage" type="u" access="read"/>
          <property name="Wattage" type="u" access="read"/>
          <property name="IncandescentEquivalent" type="u" access="read"/>
          <property name="MaxLumens" type="u" access="read"/>
          <property name="MinTemperature" type="u" access="read"/>
          <property name="MaxTemperature" type="u" access="read"/>
          <property name="ColorRenderingIndex" type="u" access="read"/>
          <property name="LampID" type="s" access="read"/>
        </interface>
        <interface name="org.allseen.LSF.LampState">
          <property name="Version" type="u" access="read"/>
          <method name="TransitionLampState">
            <arg name="Timestamp" type="t" direction="in"/>
            <arg name="NewState" type="a{sv}" direction="in"/>
            <arg name="TransitionPeriod" type="u" direction="in"/>
            <arg name="LampResponseCode" type="u" direction="out"/>
          </method>
          <method name="ApplyPulseEffect">
            <arg name="FromState" type="a{sv}" direction="in"/>
            <arg name="ToState" type="a{sv}" direction="in"/>
            <arg name="period" type="u" direction="in"/>
            <arg name="duration" type="u" direction="in"/>
            <arg name="numPulses" type="u" direction="in"/>
            <arg name="timestamp" type="t" direction="in"/>
            <arg name="LampResponseCode" type="u" direction="out"/>
          </method>
          <signal name="LampStateChanged">
            <arg name="LampID" type="s"/>
          </signal>
          <property name="OnOff" type="b" access="readwrite"/>
          <property name="Hue" type="u" access="readwrite"/>
          <property name="Saturation" type="u" access="readwrite"/>
          <property name="ColorTemp" type="u" access="readwrite"/>
          <property name="Brightness" type="u" access="readwrite"/>
        </interface>        
    """ + core.PropertiesServiceInterface.IntrospectionXml(self)

